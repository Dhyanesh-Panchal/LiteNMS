import React, { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Paper,
  Grid,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  CircularProgress,
  Alert,
  Card,
  CardContent,
  CardHeader,
} from '@mui/material';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';
import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

// Duration options in seconds
const DURATION_OPTIONS = [
  { value: 3600, label: 'Last Hour' },
  { value: 21600, label: 'Last 6 Hours' },
  { value: 86400, label: 'Last 24 Hours' },
  { value: 172800, label: 'Last 2 Days' },
  { value: 604800, label: 'Last 7 Days' },
];

const Dashboard = () => {
  const [devices, setDevices] = useState([]);
  const [selectedDevice, setSelectedDevice] = useState(null);
  const [selectedDuration, setSelectedDuration] = useState(3600); // Default to 1 hour
  const [chartData1, setChartData1] = useState([]);
  const [chartData2, setChartData2] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  // Load devices on component mount
  useEffect(() => {
    loadDevices();
  }, []);

  // Load chart data when device or duration changes
  useEffect(() => {
    if (selectedDevice) {
      loadChartData();
    }
  }, [selectedDevice, selectedDuration]);

  const loadDevices = async () => {
    try {
      setLoading(true);
      const response = await axios.get(`${API_BASE_URL}/devices`);
      setDevices(response.data.devices || []);
      
      // Auto-select the first device if available
      if (response.data.devices && response.data.devices.length > 0) {
        setSelectedDevice(response.data.devices[0].ip);
      }
      
      setLoading(false);
    } catch (err) {
      console.error('Error loading devices:', err);
      setError('Failed to load devices');
      setLoading(false);
    }
  };

  const loadChartData = async () => {
    try {
      setLoading(true);
      setError('');
      
      // Calculate from timestamp based on duration
      const toTimestamp = Math.floor(Date.now() / 1000); // Current time in seconds
      const fromTimestamp = toTimestamp - selectedDuration;
      
      // Prepare request body with the required fields: CounterID and ObjectIDs
      const baseRequestBody = {
        From: fromTimestamp,
        To: toTimestamp,
        ObjectIDs: [selectedDevice],
      };
      
      // Load data for counter ID 1
      const response1 = await axios.post(`${API_BASE_URL}/histogram`, {
        ...baseRequestBody,
        CounterID: 1,
      });
      
      // Load data for counter ID 2
      const response2 = await axios.post(`${API_BASE_URL}/histogram`, {
        ...baseRequestBody,
        CounterID: 2,
      });
      
      console.log('Counter 1 data:', response1.data);
      console.log('Counter 2 data:', response2.data);
      
      // Format data for Recharts
      setChartData1(formatChartData(response1.data));
      setChartData2(formatChartData(response2.data));
      
      setLoading(false);
    } catch (err) {
      console.error('Error loading chart data:', err);
      setError(`Failed to load monitoring data: ${err.response?.data?.error || err.message}`);
      setLoading(false);
    }
  };

  const formatChartData = (data) => {
    if (!data || !data.data || !data.data[selectedDevice] || data.data[selectedDevice].length === 0) {
      console.log('No data found for the selected device');
      return [];
    }
    
    const devicePoints = data.data[selectedDevice];
    console.log(`Found ${devicePoints.length} data points`);
    console.log('Sample data point:', devicePoints[0]);
    
    // Format the data for Recharts
    return devicePoints.map(point => {
      // Convert Unix timestamp to a JavaScript Date for formatting
      const timestamp = parseInt(point.timestamp, 10) * 1000;
      const date = new Date(timestamp);
      
      // Format the date as HH:MM:SS
      const formattedTime = date.toLocaleTimeString();
      
      return {
        time: formattedTime,     // For display on X-axis
        timestamp: timestamp,    // Raw timestamp for sorting
        value: point.value       // The actual value
      };
    })
    // Sort by timestamp to ensure chronological order
    .sort((a, b) => a.timestamp - b.timestamp);
  };

  const handleDeviceChange = (event) => {
    setSelectedDevice(event.target.value);
  };

  const handleDurationChange = (event) => {
    setSelectedDuration(event.target.value);
  };

  const formatIpAddress = (ip) => {
    return `${(ip >>> 24) & 255}.${(ip >>> 16) & 255}.${(ip >>> 8) & 255}.${ip & 255}`;
  };

  return (
    <Box sx={{ width: '100%', maxWidth: 1200, margin: '0 auto', p: 3 }}>
      <Typography variant="h4" sx={{ mb: 3 }}>
        Device Monitoring Dashboard
      </Typography>
      
      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}
      
      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2} alignItems="center">
          <Grid item xs={12} md={6}>
            <FormControl fullWidth>
              <InputLabel>Device</InputLabel>
              <Select
                value={selectedDevice || ''}
                onChange={handleDeviceChange}
                label="Device"
                disabled={loading || devices.length === 0}
              >
                {devices.map((device) => (
                  <MenuItem key={device.ip} value={device.ip}>
                    {formatIpAddress(device.ip)}
                    {device.is_provisioned ? ' (Provisioned)' : ' (Not Provisioned)'}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={6}>
            <FormControl fullWidth>
              <InputLabel>Duration</InputLabel>
              <Select
                value={selectedDuration}
                onChange={handleDurationChange}
                label="Duration"
                disabled={loading}
              >
                {DURATION_OPTIONS.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
        </Grid>
      </Paper>
      
      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', my: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <Grid container spacing={3} direction="column">
          <Grid item xs={12}>
            <Card>
              <CardHeader title="Disk Usage" />
              <CardContent sx={{ height: 400 }}>
                {chartData1.length > 0 ? (
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart
                      data={chartData1}
                      margin={{ top: 5, right: 30, left: 20, bottom: 25 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis 
                        dataKey="time" 
                        label={{ value: 'Time', position: 'insideBottomRight', offset: -10 }}
                        tick={{ fontSize: 12 }}
                        angle={-45}
                        textAnchor="end"
                      />
                      <YAxis 
                        label={{ value: 'KB', angle: -90, position: 'insideLeft' }}
                      />
                      <Tooltip formatter={(value) => [`${value} KB`, 'Disk Usage']} />
                      <Legend />
                      <Line 
                        type="monotone" 
                        dataKey="value" 
                        name="Disk Usage" 
                        stroke="#8884d8" 
                        activeDot={{ r: 6 }} 
                        dot={false}
                        strokeWidth={2}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                ) : (
                  <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                    <Typography color="text.secondary">No data available</Typography>
                  </Box>
                )}
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12}>
            <Card>
              <CardHeader title="CPU Usage" />
              <CardContent sx={{ height: 400 }}>
                {chartData2.length > 0 ? (
                  <ResponsiveContainer width="100%" height="100%">
                    <LineChart
                      data={chartData2}
                      margin={{ top: 5, right: 30, left: 20, bottom: 25 }}
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis 
                        dataKey="time" 
                        label={{ value: 'Time', position: 'insideBottomRight', offset: -10 }}
                        tick={{ fontSize: 12 }}
                        angle={-45}
                        textAnchor="end"
                      />
                      <YAxis 
                        label={{ value: '%', angle: -90, position: 'insideLeft' }}
                      />
                      <Tooltip formatter={(value) => [`${value}%`, 'CPU Usage']} />
                      <Legend />
                      <Line 
                        type="monotone" 
                        dataKey="value" 
                        name="CPU Usage" 
                        stroke="#82ca9d" 
                        activeDot={{ r: 6 }} 
                        dot={false}
                        strokeWidth={2}
                      />
                    </LineChart>
                  </ResponsiveContainer>
                ) : (
                  <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100%' }}>
                    <Typography color="text.secondary">No data available</Typography>
                  </Box>
                )}
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}
    </Box>
  );
};

export default Dashboard; 