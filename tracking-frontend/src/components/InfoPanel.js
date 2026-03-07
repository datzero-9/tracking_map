import React from 'react';

export default function InfoPanel({ lastUpdate, deviceId, currentPos, distance, duration }) {
  return (
    <div style={{ position: "absolute", bottom: "20px", left: "20px", zIndex: 1000, display: "flex", flexDirection: "column", gap: "10px" }}>
      {/* Box Tọa độ */}
      {lastUpdate && (
        <div style={{ fontSize: "13px", color: "#333", backgroundColor: "rgba(255,255,255,0.9)", padding: "10px 15px", borderRadius: "8px", boxShadow: "0 2px 10px rgba(0,0,0,0.2)" }}>
          <span style={{color: "gray"}}>Đang xem: <b>{deviceId}</b></span> <br/>
          <span style={{color: "gray"}}>Tọa độ: </span> <b>{currentPos[0].toFixed(5)}, {currentPos[1].toFixed(5)}</b> <br/>
          <span style={{color: "gray"}}>Cập nhật: </span> <b>{lastUpdate}</b>
        </div>
      )}

      {/* Box Dẫn đường */}
      {distance > 0 && (
        <div style={{ padding: "10px 15px", background: "rgba(227, 242, 253, 0.95)", borderRadius: "8px", boxShadow: "0 2px 10px rgba(33, 150, 243, 0.3)", border: "1px solid #2196F3" }}>
          <b style={{color: '#1976D2'}}>Khoảng cách: {distance} km</b> <br/>
          <b style={{color: '#E65100'}}>Thời gian xe chạy: {duration} phút</b>
        </div>
      )}
    </div>
  );
}