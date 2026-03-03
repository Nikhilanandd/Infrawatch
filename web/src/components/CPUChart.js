import React from 'react';
import { Line } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Filler,
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Title, Tooltip, Filler);

function CPUChart({ metrics }) {
  const labels = metrics.map((m) =>
    new Date(m.timestamp).toLocaleTimeString()
  );
  const data = {
    labels,
    datasets: [
      {
        label: 'CPU %',
        data: metrics.map((m) => m.cpu?.usage_percent || 0),
        borderColor: '#3b82f6',
        backgroundColor: 'rgba(59, 130, 246, 0.1)',
        fill: true,
        tension: 0.3,
        pointRadius: 0,
      },
    ],
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    scales: {
      y: {
        min: 0,
        max: 100,
        ticks: { color: '#94a3b8' },
        grid: { color: 'rgba(71, 85, 105, 0.3)' },
      },
      x: {
        ticks: { color: '#94a3b8', maxTicksLimit: 10 },
        grid: { display: false },
      },
    },
    plugins: {
      tooltip: { mode: 'index', intersect: false },
    },
  };

  return (
    <div style={{ height: '250px' }}>
      <Line data={data} options={options} />
    </div>
  );
}

export default CPUChart;
