import React from 'react';
import { Popup } from 'react-leaflet';

export default function DestinationPopup({ clickedPos, onClose, onConfirm }) {
  if (!clickedPos) return null;

  return (
    <Popup position={clickedPos} onClose={onClose}>
      <div style={{ textAlign: "center", minWidth: "150px" }}>
        <b style={{ color: "#FF5722", fontSize: "14px" }}>📍 Điểm Đích (End)</b><br />
        <span style={{ fontSize: "13px", color: "#555" }}>
          {clickedPos[0].toFixed(5)}, {clickedPos[1].toFixed(5)}
        </span><br /><br />
        <button
          onClick={(e) => {
            e.stopPropagation(); 
            e.preventDefault();
            onConfirm(clickedPos);
          }}
          style={{ 
            background: "#2196F3", color: "white", border: "none", 
            padding: "6px 15px", borderRadius: "4px", 
            cursor: "pointer", fontWeight: "bold", width: "100%" 
          }}
        >
          ✓ Chọn làm Điểm B
        </button>
      </div>
    </Popup>
  );
}