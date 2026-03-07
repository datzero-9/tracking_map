import { useState, useEffect } from 'react';
import { useMap, useMapEvents } from 'react-leaflet';
import L from 'leaflet';

// Di chuyển bản đồ 
export function MapUpdater({ center }) {
  const map = useMap();
  useEffect(() => {
    if (center && center.length === 2) {
      map.flyTo(center, map.getZoom(), { duration: 1.5 });
    }
  }, [center, map]);
  return null;
}

// Tự động zoom bản đồ 
export function RouteFitter({ route }) {
  const map = useMap();
  useEffect(() => {
    if (route && route.length > 0) {
      const bounds = L.latLngBounds(route);
      map.fitBounds(bounds, { padding: [50, 50] }); 
    }
  }, [route, map]);
  return null;
}

// Hiển thị tọa độ khi rê chuột 
export function HoverTracker() {
  const [hoverPos, setHoverPos] = useState(null);
  useMapEvents({
    mousemove(e) { setHoverPos([e.latlng.lat, e.latlng.lng]); },
    mouseout() { setHoverPos(null); }
  });

  if (!hoverPos) return null;
  return (
    <div style={{ 
      position: "absolute", top: "80px", right: "20px", zIndex: 1000, 
      background: "rgba(0, 0, 0, 0.7)", color: "white", padding: "6px 12px", 
      borderRadius: "20px", fontSize: "13px", fontWeight: "bold",
      pointerEvents: "none", boxShadow: "0 2px 5px rgba(0,0,0,0.3)" 
    }}>
      🎯 Chuột: {hoverPos[0].toFixed(5)}, {hoverPos[1].toFixed(5)}
    </div>
  );
}

// Bắt sự kiện Click 
export function MapClickHandler({ setClickedPos }) {
  useMapEvents({
    click(e) { setClickedPos([e.latlng.lat, e.latlng.lng]); }
  });
  return null;
}