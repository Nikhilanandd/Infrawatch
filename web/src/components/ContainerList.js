import React from 'react';

function ContainerList({ containers }) {
  if (!containers || containers.length === 0) {
    return (
      <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: '1rem' }}>
        No containers running
      </p>
    );
  }

  return (
    <div style={{ overflowX: 'auto' }}>
      <table className="container-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Image</th>
            <th>State</th>
            <th>CPU %</th>
            <th>Memory</th>
          </tr>
        </thead>
        <tbody>
          {containers.map((c) => (
            <tr key={c.id}>
              <td>{c.name}</td>
              <td style={{ color: 'var(--text-secondary)' }}>{c.image}</td>
              <td>
                <span className={`badge ${c.state === 'running' ? 'badge-running' : 'badge-stopped'}`}>
                  {c.state}
                </span>
              </td>
              <td>{c.cpu_percent?.toFixed(1)}%</td>
              <td>{formatBytes(c.memory_usage)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function formatBytes(bytes) {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export default ContainerList;
