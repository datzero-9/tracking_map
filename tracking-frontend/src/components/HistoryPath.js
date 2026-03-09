import React from 'react';
import { Polyline, CircleMarker, Tooltip } from 'react-leaflet';

export default function HistoryPath({ history }) {
  if (!history || history.length === 0) return null;

  return (
    <>
      <Polyline positions={history} color="red" weight={3} dashArray="5, 10" />
      
      {history.map((point, index) => {
        let color = '#FF9800'; // Màu cam cho điểm Trung gian
        let fillColor = '#FFC107';
        let label = 'Trung Gian';
        let radius = 5;

        // Điểm đầu tiên
        if (index === 0) {
          color = 'green';
          fillColor = 'green';
          label = 'Bắt Đầu';
          radius = 8;
        } 
        // Điểm cuối cùng
        else if (index === history.length - 1) {
          color = 'black';
          fillColor = 'black';
          label = 'Kết Thúc';
          radius = 8;
        }

        return (
          <CircleMarker 
            key={index} 
            center={point} 
            radius={radius} 
            pathOptions={{ color: color, fillColor: fillColor, fillOpacity: 0.8 }}
          >
            <Tooltip direction="top">
              <b style={{ color: color === 'black' ? '#333' : color }}>{label}</b><br />
              {point[0].toFixed(5)}, {point[1].toFixed(5)}
            </Tooltip>
          </CircleMarker>
        );
      })}
    </>
  );
}