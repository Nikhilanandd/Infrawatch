import React from 'react';

function AlertHistory({ alerts }) {
  if (!alerts || alerts.length === 0) {
    return (
      <p style={{ color: 'var(--text-secondary)', textAlign: 'center', padding: '1rem' }}>
        No alerts
      </p>
    );
  }

  return (
    <div style={{ maxHeight: '300px', overflowY: 'auto' }}>
      {alerts.map((alert, idx) => (
        <div key={alert.id || idx} className={`alert-item ${alert.severity}`}>
          <div className="alert-time">
            {new Date(alert.fired_at).toLocaleString()} — {alert.severity.toUpperCase()}
          </div>
          <div className="alert-msg">{alert.message}</div>
        </div>
      ))}
    </div>
  );
}

export default AlertHistory;
