import React from 'react';
import { Bar } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
} from 'chart.js';

ChartJS.register(CategoryScale, LinearScale, BarElement, Title, Tooltip);

function DiskChart({ metric }) {
  if (!metric?.disk?.partitions?.length) {
    return <p style={{ color: 'var(--text-secondary)', textAlign: 'center' }}>No disk data</p>;
  }

  const partitions = metric.disk.partitions;

  const data = {
    labels: partitions.map((p) => p.mountpoint),
    datasets: [
      {
        label: 'Used %',
        data: partitions.map((p) => p.usage_percent),
        backgroundColor: partitions.map((p) =>
          p.usage_percent > 85
            ? 'rgba(239, 68, 68, 0.7)'
            : p.usage_percent > 70
            ? 'rgba(245, 158, 11, 0.7)'
            : 'rgba(34, 197, 94, 0.7)'
        ),
        borderRadius: 6,
      },
    ],
  };

  const options = {
    responsive: true,
    maintainAspectRatio: false,
    indexAxis: 'y',
    scales: {
      x: {
        min: 0,
        max: 100,
        ticks: { color: '#94a3b8' },
        grid: { color: 'rgba(71, 85, 105, 0.3)' },
      },
      y: {
        ticks: { color: '#94a3b8' },
        grid: { display: false },
      },
    },
    plugins: {
      tooltip: {
        callbacks: {
          afterLabel: (ctx) => {
            const p = partitions[ctx.dataIndex];
            return `${formatBytes(p.used_bytes)} / ${formatBytes(p.total_bytes)}`;
          },
        },
      },
    },
  };

  return (
    <div style={{ height: '250px' }}>
      <Bar data={data} options={options} />
    </div>
  );
}

function formatBytes(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export default DiskChart;
