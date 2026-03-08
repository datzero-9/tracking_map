import React, { useState, useEffect, useRef } from 'react';
import { MapContainer, TileLayer, Polyline, Marker, Popup, CircleMarker, Tooltip } from 'react-leaflet';
import axios from 'axios';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';
import ControlPanel from './components/ControlPanel';
import InfoPanel from './components/InfoPanel';
import { MapUpdater, RouteFitter, HoverTracker, MapClickHandler } from './components/MapEvents';

// Icon
delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
});

const API_BASE = "http://localhost:8080/api/v1";

function App() {
  const [deviceId, setDeviceId] = useState("car_01");
  const [deviceList, setDeviceList] = useState([]);
  const [currentPos, setCurrentPos] = useState([16.047, 108.206]);
  const [lastUpdate, setLastUpdate] = useState("");

  const today = new Date().toISOString().split('T')[0];
  const [history, setHistory] = useState([]);
  const [fromTime, setFromTime] = useState(`${today}T00:00`);
  const [toTime, setToTime] = useState(`${today}T23:59`);

  const [startPoint, setStartPoint] = useState("");
  const [endPoint, setEndPoint] = useState("");
  const [route, setRoute] = useState([]);
  const [distance, setDistance] = useState(0);
  const [duration, setDuration] = useState(0);

  const [clickedPos, setClickedPos] = useState(null);
  const [destPos, setDestPos] = useState(null);

  const markerRef = useRef(null);

  useEffect(() => {
    const fetchDevices = async () => {
      try {
        const res = await axios.get(`${API_BASE}/devices`);
        if (res.data) {
          setDeviceList(res.data);
        }
      } catch (e) { console.error("Lỗi lấy danh sách thiết bị"); }
    };
    fetchDevices();
  }, []);

  // Tìm thiết bị và cập nhật vị trí

  const handleFindDevice = async (targetId) => {
    let idToFind = deviceId;
    if (typeof targetId === 'string') {
      idToFind = targetId;
    }

    if (!idToFind) {
      alert("Vui lòng nhập ID xe!");
      return;
    }

    try {
      const cleanId = idToFind.trim();
      const res = await axios.get(`${API_BASE}/devices/${cleanId}/latest`);
      const { latitude, longitude, timestamp } = res.data;
      setCurrentPos([latitude, longitude]);
      setLastUpdate(new Date(timestamp).toLocaleString('vi-VN'));
      if (markerRef.current) markerRef.current.openPopup();
    } catch (e) {
      alert(`Không tìm thấy dữ liệu cho thiết bị: "${idToFind}"\nHãy chắc chắn bạn đã chạy xe này!`);
    }
  };


  // Xử lý thay đổi thiết bị
  const onDeviceChange = (e) => {
    const newValue = e.target.value;
    setDeviceId(newValue);
    if (deviceList.includes(newValue)) {
      handleFindDevice(newValue);
    }
  };


  // Lấy vị trí hiện tại và lưu vào DB
  const handleUseCurrentLocation = () => {
    navigator.geolocation.getCurrentPosition(async (pos) => {
      const { latitude, longitude } = pos.coords;
      const currentTime = new Date().toISOString();
      setCurrentPos([latitude, longitude]);
      setLastUpdate(new Date().toLocaleString('vi-VN'));
      setStartPoint(`${latitude.toFixed(6)}, ${longitude.toFixed(6)}`);
      setRoute([]); setDistance(0); setDuration(0);
      try {
        await axios.post(`${API_BASE}/locations`, { device_id: deviceId, latitude, longitude, timestamp: currentTime });
        if (!deviceList.includes(deviceId)) {
          setDeviceList([...deviceList, deviceId]);
        }
        alert(`Đã thêm thiết bị mới: ${deviceId} và lưu vị trí hiện tại vào DB!`);

      } catch (err) {
        console.error("Lỗi lưu DB", err);
      }
      if (markerRef.current) {
        markerRef.current.openPopup();
      }
    }, (err) => alert("Vui lòng bật quyền truy cập GPS!"), { enableHighAccuracy: true });
  };


  // Xử lý xem lịch sử
  const handleToggleHistory = async () => {
    if (history.length > 0) {
      setHistory([]);
      return;
    }
    try {
      const res = await axios.get(`${API_BASE}/devices/${deviceId}/history`, {
        params: { from_time: new Date(fromTime).toISOString(), to_time: new Date(toTime).toISOString() }
      });
      if (res.data.length > 0) {
        const points = res.data.map(loc => [loc.latitude, loc.longitude]);
        setHistory(points);
        setCurrentPos(points[points.length - 1]);
      } else {
        alert("Không có lịch sử trong khoảng thời gian này!"); setHistory([]);
      }
    } catch (e) {
      alert("Lỗi lấy lịch sử!");
    }
  };


  // Xử lý vẽ tuyến đường
  const handleRouting = async () => {
    try {
      const startParts = startPoint.split(',');
      const endParts = endPoint.split(',');
      if (startParts.length !== 2 || endParts.length !== 2) {
        alert("Vui lòng nhập đầy đủ Tọa độ Điểm A và Điểm B");
        return;
      }
      const sLat = parseFloat(startParts[0].trim());
      const sLon = parseFloat(startParts[1].trim());
      const eLat = parseFloat(endParts[0].trim());
      const eLon = parseFloat(endParts[1].trim());
      if (isNaN(sLat) || isNaN(sLon) || isNaN(eLat) || isNaN(eLon)) {
        alert("Tọa độ phải là số hợp lệ!");
        return;
      }

      const resRoute = await axios.get(`${API_BASE}/route`, {
        params: { start_lat: sLat, start_lon: sLon, end_lat: eLat, end_lon: eLon }
      });
      const routeData = resRoute.data.routes[0];
      const coords = routeData.geometry.coordinates.map(p => [p[1], p[0]]);
      setRoute(coords);
      setDistance((routeData.distance / 1000).toFixed(2));
      setDuration(Math.round(routeData.duration / 60));
    } catch (e) {
      alert("Lỗi: Không tìm thấy lộ trình kết nối 2 điểm này!");
    }
  };

  return (
    <div style={{ height: "100vh", position: "relative", fontFamily: "sans-serif" }}>

      <ControlPanel
        deviceId={deviceId} deviceList={deviceList} onDeviceChange={onDeviceChange} handleFindDevice={handleFindDevice}
        fromTime={fromTime} setFromTime={setFromTime} toTime={toTime} setToTime={setToTime} handleToggleHistory={handleToggleHistory} history={history}
        startPoint={startPoint} setStartPoint={setStartPoint} endPoint={endPoint} setEndPoint={setEndPoint}
        handleUseCurrentLocation={handleUseCurrentLocation} handleRouting={handleRouting}
        setRoute={setRoute} setDistance={setDistance} setDuration={setDuration} setDestPos={setDestPos}
      />

      <InfoPanel
        lastUpdate={lastUpdate} deviceId={deviceId} currentPos={currentPos} distance={distance} duration={duration}
      />

      <MapContainer center={currentPos} zoom={15} style={{ height: "100%", width: "100%", cursor: "crosshair" }}>
        {/* xử lý các sự trên bản đồ */}
        <MapUpdater center={currentPos} />
        <HoverTracker />
        <MapClickHandler setClickedPos={setClickedPos} />
        <RouteFitter route={route} />
        {clickedPos && (
          <Popup position={clickedPos} onClose={() => setClickedPos(null)}>
            <div style={{ textAlign: "center", minWidth: "150px" }}>
              <b style={{ color: "#FF5722", fontSize: "14px" }}>📍 Điểm Đích (End)</b><br />
              <span style={{ fontSize: "13px", color: "#555" }}>{clickedPos[0].toFixed(5)}, {clickedPos[1].toFixed(5)}</span><br /><br />
              <button
                onClick={(e) => {
                  e.stopPropagation(); e.preventDefault();
                  setEndPoint(`${clickedPos[0].toFixed(5)}, ${clickedPos[1].toFixed(5)}`);
                  setRoute([]); setDistance(0); setDuration(0);
                  setDestPos([clickedPos[0], clickedPos[1]]);
                  setClickedPos(null);
                }}
                style={{ background: "#2196F3", color: "white", border: "none", padding: "6px 15px", borderRadius: "4px", cursor: "pointer", fontWeight: "bold", width: "100%" }}
              >
                ✓ Chọn làm Điểm B
              </button>
            </div>
          </Popup>
        )}

        {destPos && (
          <CircleMarker center={destPos} radius={10} pathOptions={{ color: '#E65100', fillColor: '#FF9800', fillOpacity: 1 }}>
            <Tooltip direction="top" permanent offset={[0, -10]}><b style={{ color: '#E65100' }}>📍 Điểm B</b></Tooltip>
          </CircleMarker>
        )}

        {history.length > 0 && (
          <>
            <Polyline positions={history} color="red" weight={3} dashArray="5, 10" />
            {history.map((point, index) => {
              if (index === 0) return (<CircleMarker key={index} center={point} radius={8} pathOptions={{ color: 'green', fillColor: 'green', fillOpacity: 1 }}><Tooltip direction="top"><b>Bắt Đầu</b><br />{point[0]}, {point[1]}</Tooltip></CircleMarker>);
              if (index === history.length - 1) return (<CircleMarker key={index} center={point} radius={8} pathOptions={{ color: 'black', fillColor: 'black', fillOpacity: 1 }}><Tooltip direction="top"><b>Kết Thúc</b><br />{point[0]}, {point[1]}</Tooltip></CircleMarker>);
              return (<CircleMarker key={index} center={point} radius={5} pathOptions={{ color: '#FF9800', fillColor: '#FFC107', fillOpacity: 0.8 }}><Tooltip direction="top"><b style={{ color: '#FF9800' }}>Trung Gian</b><br />{point[0]}, {point[1]}</Tooltip></CircleMarker>);
            })}
          </>
        )}

        {route.length > 0 && <Polyline positions={route} color="blue" weight={6} opacity={0.6} />}

        <Marker position={currentPos} ref={markerRef}>
          <Popup>
            <div style={{ textAlign: "center" }}>
              <b style={{ fontSize: "16px", color: "#4CAF50" }}>{deviceId} (Điểm A)</b> <br />
              <hr style={{ margin: "5px 0" }} />
              <small>Tọa độ: {currentPos[0].toFixed(5)}, {currentPos[1].toFixed(5)}</small><br />
              <small style={{ color: "gray" }}>{lastUpdate || "Vị trí hiện tại"}</small>
            </div>
          </Popup>
        </Marker>
      </MapContainer>
    </div>
  );
}

export default App;