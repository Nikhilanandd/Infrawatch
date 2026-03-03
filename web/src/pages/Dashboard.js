import React, { useState, useEffect, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../hooks/useAuth';
import { useWebSocket } from '../hooks/useWebSocket';
import { getNodes, getMetrics, getAlerts } from '../services/api';
import CPUChart from '../components/CPUChart';
import MemoryChart from '../components/MemoryChart';
import DiskChart from '../components/DiskChart';
import ContainerList from '../components/ContainerList';
import AlertHistory from '../components/AlertHistory';

function Dashboard() {
  const { role, logout } = useAuth();
  const navigate = useNavigate();

  const [nodes, setNodes] = useState([]);
  const [selectedNode, setSelectedNode] = useState(null);
  const [metrics, setMetrics] = useState([]);
  const [alerts, setAlerts] = useState([]);
  const [latestMetric, setLatestMetric] = useState(null);

  const wsUrl = `${window.location.protocol === 'https:' ? 'wss' : 'ws'}://${window.location.host}/api/v1/ws`;
  const { lastMessage, isConnected } = useWebSocket(wsUrl);

  // Fetch nodes
  const fetchNodes = useCallback(async () => {
    try {
      const res = await getNodes();
      setNodes(res.data || []);
      if (!selectedNode && res.data?.length > 0) {
        setSelectedNode(res.data[0].id);
      }
    } catch (err) {
      console.error('Failed to fetch nodes:', err);
    }
  }, [selectedNode]);

  // Fetch metrics for selected node
  const fetchMetrics = useCallback(async () => {
    if (!selectedNode) return;
    try {
      const from = new Date(Date.now() - 60 * 60 * 1000).toISOString();
      const to = new Date().toISOString();
      const res = await getMetrics(selectedNode, from, to, 200);
      const data = (res.data || []).reverse(); // oldest first
      setMetrics(data);
      if (data.length > 0) {
        setLatestMetric(data[data.length - 1]);
      }
    } catch (err) {
      console.error('Failed to fetch metrics:', err);
    }
  }, [selectedNode]);

  // Fetch alerts
  const fetchAlerts = useCallback(async () => {
    try {
      const res = await getAlerts(selectedNode, 20);
      setAlerts(res.data || []);
    } catch (err) {
      console.error('Failed to fetch alerts:', err);
    }
  }, [selectedNode]);

  useEffect(() => {
    fetchNodes();
    const interval = setInterval(fetchNodes, 15000);
    return () => clearInterval(interval);
  }, [fetchNodes]);

  useEffect(() => {
    fetchMetrics();
    fetchAlerts();
    const interval = setInterval(() => {
      fetchMetrics();
      fetchAlerts();
    }, 10000);
    return () => clearInterval(interval);
  }, [fetchMetrics, fetchAlerts]);

  // Handle WebSocket live updates
  useEffect(() => {
    if (!lastMessage) return;
    if (lastMessage.type === 'metric' && lastMessage.payload?.node_id === selectedNode) {
      setMetrics((prev) => [...prev.slice(-199), lastMessage.payload]);
      setLatestMetric(lastMessage.payload);
    }
    if (lastMessage.type === 'alert') {
      setAlerts((prev) => [lastMessage.payload, ...prev].slice(0, 20));
    }
  }, [lastMessage, selectedNode]);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  return (
    <div className="dashboard">
      <nav className="navbar">
        <h1>InfraWatch</h1>
        <div className="user-info">
          <span style={{ color: isConnected ? 'var(--success)' : 'var(--danger)' }}>
            ● {isConnected ? 'Live' : 'Disconnected'}
          </span>
          <span>Role: {role}</span>
          <button className="btn-sm" onClick={handleLogout}>Logout</button>
        </div>
      </nav>

      <div className="dashboard-content">
        {/* Top stats */}
        <div className="grid grid-3" style={{ marginBottom: '1.5rem' }}>
          <div className="card">
            <h2>CPU Usage</h2>
            <div className="stat-value" style={{ color: (latestMetric?.cpu?.usage_percent || 0) > 80 ? 'var(--danger)' : 'var(--success)' }}>
              {(latestMetric?.cpu?.usage_percent || 0).toFixed(1)}%
            </div>
          </div>
          <div className="card">
            <h2>Memory Usage</h2>
            <div className="stat-value" style={{ color: (latestMetric?.memory?.usage_percent || 0) > 80 ? 'var(--danger)' : 'var(--success)' }}>
              {(latestMetric?.memory?.usage_percent || 0).toFixed(1)}%
            </div>
          </div>
          <div className="card">
            <h2>Uptime</h2>
            <div className="stat-value">
              {formatUptime(latestMetric?.uptime_seconds || 0)}
            </div>
          </div>
        </div>

        <div className="grid grid-2">
          {/* Node List */}
          <div className="card">
            <h2>Nodes ({nodes.length})</h2>
            <ul className="node-list">
              {nodes.map((node) => (
                <li
                  key={node.id}
                  className={`node-item ${selectedNode === node.id ? 'active' : ''}`}
                  onClick={() => setSelectedNode(node.id)}
                >
                  <span>
                    <span className={`status-dot status-${node.status}`} />
                    {node.hostname}
                  </span>
                  <span style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                    {node.id}
                  </span>
                </li>
              ))}
              {nodes.length === 0 && (
                <li style={{ color: 'var(--text-secondary)', padding: '1rem', textAlign: 'center' }}>
                  No agents connected yet
                </li>
              )}
            </ul>
          </div>

          {/* Alert History */}
          <div className="card">
            <h2>Alert History</h2>
            <AlertHistory alerts={alerts} />
          </div>
        </div>

        {/* Charts */}
        <div className="grid grid-2" style={{ marginTop: '1.5rem' }}>
          <div className="card">
            <h2>CPU Usage Over Time</h2>
            <CPUChart metrics={metrics} />
          </div>
          <div className="card">
            <h2>Memory Usage Over Time</h2>
            <MemoryChart metrics={metrics} />
          </div>
        </div>

        <div className="grid grid-2" style={{ marginTop: '1.5rem' }}>
          <div className="card">
            <h2>Disk Usage</h2>
            <DiskChart metric={latestMetric} />
          </div>
          <div className="card">
            <h2>Containers</h2>
            <ContainerList containers={latestMetric?.containers || []} />
          </div>
        </div>
      </div>
    </div>
  );
}

function formatUptime(seconds) {
  const d = Math.floor(seconds / 86400);
  const h = Math.floor((seconds % 86400) / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  if (d > 0) return `${d}d ${h}h`;
  if (h > 0) return `${h}h ${m}m`;
  return `${m}m`;
}

export default Dashboard;
