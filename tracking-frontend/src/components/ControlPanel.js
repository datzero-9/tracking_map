import React from 'react';

export default function ControlPanel(props) {
  const {
    deviceId, deviceList, onDeviceChange, handleFindDevice,
    fromTime, setFromTime, toTime, setToTime, handleToggleHistory, history,
    startPoint, setStartPoint, endPoint, setEndPoint, handleUseCurrentLocation, handleRouting,
    setRoute, setDistance, setDuration, setDestPos
  } = props;

  return (
    <div style={{ 
      position: "absolute", top: "10px", left: "50%", transform: "translateX(-50%)", 
      zIndex: 1000, background: "rgba(255, 255, 255, 0.95)", 
      padding: "10px 20px", borderRadius: "8px", boxShadow: "0 4px 15px rgba(0,0,0,0.2)",
      display: "flex", flexWrap: "wrap", gap: "20px", alignItems: "center", justifyContent: "center",
      width: "90%", maxWidth: "1200px"
    }}>
      
      {/* Tìm thiết bị  */}
      <div style={{ display: "flex", gap: "8px", alignItems: "center", borderRight: "1px solid #ddd", paddingRight: "15px" }}>
        <b>Xe:</b>
        <select value={deviceList.includes(deviceId) ? deviceId : ""} onChange={onDeviceChange} style={{ padding: "6px", borderRadius: "4px", border: "1px solid #ccc", outline: "none", cursor: "pointer", background: "white" }}>
          <option value="" disabled>-- Menu Xe --</option>
          {deviceList.map((id, index) => <option key={index} value={id}>{id}</option>)}
        </select>
        <input value={deviceId} onChange={onDeviceChange} placeholder="Hoặc gõ ID..." title="Gõ ID mới nếu xe chưa có trong Menu" style={{ width: "95px", padding: "6px", borderRadius: "4px", border: "1px solid #ccc", outline: "none" }} />
        <button onClick={handleFindDevice} style={{ background: "#4CAF50", color: "white", border: "none", padding: "7px 12px", cursor: "pointer", borderRadius: "4px", fontWeight: "bold" }}>Định vị</button>
      </div>

      {/* Lịch sử */}
      <div style={{ display: "flex", gap: "8px", alignItems: "center", borderRight: "1px solid #ddd", paddingRight: "15px" }}>
          <b>Lịch sử:</b>
          <input type="datetime-local" value={fromTime} onChange={e => setFromTime(e.target.value)} style={{ padding: "5px", borderRadius: "4px", border: "1px solid #ccc" }}/>
          <span>-</span>
          <input type="datetime-local" value={toTime} onChange={e => setToTime(e.target.value)} style={{ padding: "5px", borderRadius: "4px", border: "1px solid #ccc" }}/>
          <button onClick={handleToggleHistory} style={{ background: history.length > 0 ? "#757575" : "#f44336", color: "white", border: "none", padding: "7px 12px", cursor: "pointer", borderRadius: "4px", fontWeight: "bold", width: "60px" }}>
            {history.length > 0 ? "Ẩn" : "Xem"}
          </button>
      </div>

      {/* Dẫn đường  */}
      <div style={{ display: "flex", gap: "8px", alignItems: "center" }}>
        <b>Dẫn đường:</b>
        <input 
          value={startPoint} 
          onChange={e => { setStartPoint(e.target.value); setRoute([]); setDistance(0); setDuration(0); }} 
          placeholder="A (Vĩ, Kinh)" style={{ width: "150px", padding: "6px", borderRadius: "4px", border: "1px solid #ccc" }}
        />
        <button onClick={handleUseCurrentLocation} title="Lấy GPS hiện tại" style={{ padding: "5px 8px", cursor: "pointer", borderRadius: "4px", border: "1px solid #ccc", background: "#f0f0f0" }}>📍</button>
        <input 
          value={endPoint} 
          onChange={e => { setEndPoint(e.target.value); setRoute([]); setDistance(0); setDuration(0); setDestPos(null); }} 
          placeholder="B (Hover & Click bản đồ)" style={{ width: "160px", padding: "6px", borderRadius: "4px", border: "2px solid #2196F3", outline: "none" }} title="Rê chuột vào bản đồ để xem tọa độ, click để chọn"
        />
        <button onClick={handleRouting} style={{ background: "#2196F3", color: "white", border: "none", padding: "7px 12px", cursor: "pointer", borderRadius: "4px", fontWeight: "bold" }}>Vẽ tuyến</button>
      </div>
    </div>
  );
}