import React, { useState, useEffect, useRef } from 'react';
// 👇 1. Thêm Polygon vào import
import { MapContainer, TileLayer, Polyline, Marker, Popup, CircleMarker, Tooltip, Polygon } from 'react-leaflet';
import axios from 'axios';
import 'leaflet/dist/leaflet.css';
import L from 'leaflet';

import ControlPanel from './components/ControlPanel';
import InfoPanel from './components/InfoPanel';
import { MapUpdater, RouteFitter, HoverTracker, MapClickHandler } from './components/MapEvents';
import HistoryPath from './components/HistoryPath';
import DestinationPopup from './components/DestinationPopup';

delete L.Icon.Default.prototype._getIconUrl;
L.Icon.Default.mergeOptions({
  iconRetinaUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon-2x.png',
  iconUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-icon.png',
  shadowUrl: 'https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.7.1/images/marker-shadow.png',
});

const API_BASE = "http://localhost:8080/api/v1";

function App() {

  const today = new Date().toISOString().split('T')[0];

  const [deviceId, setDeviceId] = useState("car_01");
  const [deviceList, setDeviceList] = useState([]);
  const [currentPos, setCurrentPos] = useState([15.983176, 108.254083]);
  const [lastUpdate, setLastUpdate] = useState("");

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

  // 👇 2. STATE VÀ USE-EFFECT ĐỂ KÉO FILE BIÊN GIỚI TỪ BACKEND
  const [maskPositions, setMaskPositions] = useState([]);

  useEffect(() => {
    fetch('http://localhost:8080/api/boundary')
      .then(response => {
        if (!response.ok) throw new Error("Không lấy được viền biên giới");
        return response.json();
      })
      .then(data => {
        const worldBounds = [[90, -180], [90, 180], [-90, 180], [-90, -180]];
        const geometry = data.features[0].geometry;
        let holes = [];

        if (geometry.type === "Polygon") {
          holes.push(geometry.coordinates[0].map(c => [c[1], c[0]]));
        } else if (geometry.type === "MultiPolygon") {
          geometry.coordinates.forEach(poly => {
            holes.push(poly[0].map(c => [c[1], c[0]]));
          });
        }
        
        setMaskPositions([worldBounds, ...holes]);
      })
      .catch(error => console.error("Lỗi kéo mặt nạ:", error));
  }, []);
  // 👆 ----------------------------------------------------- 👆

  useEffect(() => {
    const fetchDevices = async () => {
      try {
        const res = await axios.get(`${API_BASE}/devices`);
        if (res.data) setDeviceList(res.data);
      } catch (e) { console.error("Lỗi lấy danh sách thiết bị"); }
    };
    fetchDevices();
  }, []);

  const handleFindDevice = async (targetId) => {
    let idToFind = typeof targetId === 'string' ? targetId : deviceId;
    if (!idToFind) return alert("Vui lòng nhập ID xe!");

    try {
      const res = await axios.get(`${API_BASE}/devices/${idToFind.trim()}/latest`);
      const { latitude, longitude, timestamp } = res.data;
      setCurrentPos([latitude, longitude]);
      setLastUpdate(new Date(timestamp).toLocaleString('vi-VN'));
      if (markerRef.current) markerRef.current.openPopup();
    } catch (e) {
      alert(`Không tìm thấy dữ liệu cho thiết bị: "${idToFind}"\nHãy chắc chắn bạn đã chạy xe này!`);
    }
  };

  const onDeviceChange = (e) => {
    const newValue = e.target.value;
    setDeviceId(newValue);
    if (deviceList.includes(newValue)) handleFindDevice(newValue);
  };

  const handleUseCurrentLocation = () => {
    if (!navigator.geolocation) return alert("Trình duyệt không hỗ trợ GPS!");

    navigator.geolocation.getCurrentPosition(
      async (pos) => {
        const { latitude, longitude } = pos.coords;
        const currentTime = new Date().toISOString();

        setCurrentPos([latitude, longitude]);
        setLastUpdate(new Date().toLocaleString("vi-VN"));
        setStartPoint(`${latitude.toFixed(6)}, ${longitude.toFixed(6)}`);
        setRoute([]); setDistance(0); setDuration(0);

        try {
          await axios.post(`${API_BASE}/locations`, {
            device_id: deviceId, latitude, longitude, timestamp: currentTime,
          });
          setDeviceList(prev => !prev.includes(deviceId) ? [...prev, deviceId] : prev);
        } catch (err) { console.error("Lỗi lưu DB:", err); }

        if (markerRef.current) markerRef.current.openPopup();
        alert(`Đã thêm thiết bị ${deviceId} và lưu vị trí hiện tại!`);
      },
      (err) => alert("Vui lòng bật quyền truy cập GPS!"),
      { enableHighAccuracy: true, timeout: 10000, maximumAge: 0 }
    );
  };

  const handleToggleHistory = async () => {
    if (history.length > 0) return setHistory([]);

    try {
      const res = await axios.get(`${API_BASE}/devices/${deviceId}/history`, {
        params: { from_time: new Date(fromTime).toISOString(), to_time: new Date(toTime).toISOString() }
      });
      if (res.data.length > 0) {
        const points = res.data.map(loc => [loc.latitude, loc.longitude]);
        setHistory(points);
        setCurrentPos(points[points.length - 1]);
      } else {
        alert("Không có lịch sử trong khoảng thời gian này!");
        setHistory([]);
      }
    } catch (e) { alert("Lỗi lấy lịch sử!"); }
  };

  const handleRouting = async () => {
    try {
      const startParts = startPoint.split(',');
      const endParts = endPoint.split(',');
      if (startParts.length !== 2 || endParts.length !== 2)
        return alert("Vui lòng nhập đầy đủ Tọa độ Điểm A và Điểm B");

      const sLat = parseFloat(startParts[0].trim());
      const sLon = parseFloat(startParts[1].trim());
      const eLat = parseFloat(endParts[0].trim());
      const eLon = parseFloat(endParts[1].trim());

      if (isNaN(sLat) || isNaN(sLon) || isNaN(eLat) || isNaN(eLon)) return alert("Tọa độ phải là số hợp lệ!");

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

  const handleConfirmDestination = (pos) => {
    setEndPoint(`${pos[0].toFixed(5)}, ${pos[1].toFixed(5)}`);
    setRoute([]);
    setDistance(0);
    setDuration(0);
    setDestPos([pos[0], pos[1]]);
    setClickedPos(null);
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

      <InfoPanel lastUpdate={lastUpdate} deviceId={deviceId} currentPos={currentPos} distance={distance} duration={duration} />

      {/* 👇 Đã thêm backgroundColor: "#f4f4f4" để đồng bộ màu viền */}
      <MapContainer
        center={currentPos}
        zoom={12}
        maxZoom={15}
        minZoom={6}
        style={{ height: "100vh", width: "100%", backgroundColor: "#f4f4f4" }}>

        <TileLayer
          url="http://localhost:8080/tiles/{z}/{x}/{y}.png"
          attribution='&copy; Bản đồ Offline (Google Tiles)'
          minNativeZoom={10} 
          maxNativeZoom={12} 
        />

        {/* {maskPositions.length > 0 && (
          <Polygon 
            positions={maskPositions} 
            pathOptions={{ 
              color: 'transparent', 
              fillColor: '#f4f4f4', 
              fillOpacity: 1 
            }} 
          />
        )} */}

        <MapUpdater center={currentPos} />
        <HoverTracker />
        <MapClickHandler setClickedPos={setClickedPos} />
        <RouteFitter route={route} />

        <DestinationPopup clickedPos={clickedPos}
          onClose={() => setClickedPos(null)}
          onConfirm={handleConfirmDestination} />

        {destPos && (
          <CircleMarker
            center={destPos}
            radius={10}
            pathOptions={{ color: '#E65100', fillColor: '#FF9800', fillOpacity: 1 }}>
            <Tooltip direction="top" permanent offset={[0, -10]}><b style={{ color: '#E65100' }}>📍 Điểm B</b></Tooltip>
          </CircleMarker>
        )}

        <HistoryPath history={history} />

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