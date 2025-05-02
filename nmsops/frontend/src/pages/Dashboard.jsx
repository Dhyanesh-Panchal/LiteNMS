import React, { useState, useEffect, useRef } from 'react';
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
  Switch,
  FormControlLabel,
  Button,
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
import RefreshIcon from '@mui/icons-material/Refresh';

const API_BASE_URL = 'http://localhost:8080/api';
const REFRESH_INTERVAL = 5000; // 5 seconds

// Duration options in seconds
const DURATION_OPTIONS = [
  { value: 3600, label: 'Last Hour', interval: 1 }, // 1 second interval
  { value: 21600, label: 'Last 6 Hours', interval: 5 }, // 5 seconds interval
  { value: 86400, label: 'Last 24 Hours', interval: 10 }, // 10 seconds interval
  { value: 172800, label: 'Last 2 Days', interval: 30 }, // 30 seconds interval
  { value: 604800, label: 'Last 7 Days', interval: 30 }, // 30 seconds interval
];

// Aggregation options
const AGGREGATION_OPTIONS = [
  { value: 'avg', label: 'Average' },
  { value: 'min', label: 'Minimum' },
  { value: 'max', label: 'Maximum' },
  { value: 'sum', label: 'Sum' },
  { value: 'count', label: 'Count' },
];

const Dashboard = () => {
  const [devices, setDevices] = useState([]);
  const [selectedDevice, setSelectedDevice] = useState('');
  const [selectedDuration, setSelectedDuration] = useState(3600); // Default to 1 hour
  const [selectedAggregation, setSelectedAggregation] = useState('avg');
  const [chartData1, setChartData1] = useState([]);
  const [chartData2, setChartData2] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const chartRef1 = useRef(null);
  const chartRef2 = useRef(null);

  // Load devices on component mount
  useEffect(() => {
    loadDevices();
  }, []);

  // Load chart data when device, duration, or aggregation changes
  useEffect(() => {
    if (selectedDevice) {
      loadChartData();
    }
  }, [selectedDevice, selectedDuration, selectedAggregation]);

  const loadDevices = async () => {
    try {
      setLoading(true);
      const response = await axios.get(`${API_BASE_URL}/devices`);
      setDevices(response.data.devices || []);
      setLoading(false);
    } catch (err) {
      console.error('Error loading devices:', err);
      setError('Failed to load devices');
      setLoading(false);
    }
  };

  const loadChartData = async () => {
    if (!selectedDevice) return;

    try {
      setLoading(true);
      setError('');

      // Calculate timestamps based on current time and selected duration
      const toTimestamp = Math.floor(Date.now() / 1000);
      const fromTimestamp = toTimestamp - selectedDuration;

      // Get the interval based on selected duration
      const selectedOption = DURATION_OPTIONS.find(option => option.value === selectedDuration);
      const interval = selectedOption ? selectedOption.interval : 1;

      // Prepare request body
      const requestBody = {
        from: fromTimestamp,
        to: toTimestamp,
        object_ids: [selectedDevice],
        vertical_aggregation: selectedAggregation,
        horizontal_aggregation: selectedAggregation,
        interval: interval
      };

      // Load data for both charts
      const [response1, response2] = await Promise.all([
        axios.post(`${API_BASE_URL}/query`, { ...requestBody, counter_id: 1 }),
        axios.post(`${API_BASE_URL}/query`, { ...requestBody, counter_id: 2 })
      ]);

      // Format and set chart data
      setChartData1(formatChartData(response1.data));
      setChartData2(formatChartData(response2.data));
      setLoading(false);
    } catch (err) {
      console.error('Error loading chart data:', err);
      setError('Failed to load monitoring data');
      setLoading(false);
    }
  };

  const formatChartData = (data) => {
    if (!data || !Array.isArray(data) || data.length === 0) {
      console.log('No data found in response');
      return [];
    }
    
    // Format the data for Recharts
    return data.map(point => {
      const timestamp = point.timestamp * 1000; // Convert to milliseconds
      const date = new Date(timestamp);
      return {
        time: date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' }),
        timestamp: timestamp,
        value: point.value
      };
    })
    // Sort by timestamp to ensure chronological order
    .sort((a, b) => a.timestamp - b.timestamp);
  };

  const handleRefresh = () => {
    loadChartData();
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
          <Grid item xs={12} md={3}>
            <FormControl fullWidth>
              <InputLabel>Device</InputLabel>
              <Select
                value={selectedDevice}
                onChange={(e) => setSelectedDevice(e.target.value)}
                label="Device"
                disabled={loading}
              >
                {devices.map((device) => (
                  <MenuItem key={device.ip} value={device.ip}>
                    {device.ip}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth>
              <InputLabel>Duration</InputLabel>
              <Select
                value={selectedDuration}
                onChange={(e) => setSelectedDuration(e.target.value)}
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
          <Grid item xs={12} md={3}>
            <FormControl fullWidth>
              <InputLabel>Aggregation</InputLabel>
              <Select
                value={selectedAggregation}
                onChange={(e) => setSelectedAggregation(e.target.value)}
                label="Aggregation"
                disabled={loading}
              >
                {AGGREGATION_OPTIONS.map((option) => (
                  <MenuItem key={option.value} value={option.value}>
                    {option.label}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <Button
              variant="contained"
              onClick={handleRefresh}
              disabled={loading || !selectedDevice}
              startIcon={<RefreshIcon />}
              fullWidth
            >
              Refresh Charts
            </Button>
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
                      ref={chartRef1}
                      data={chartData1}
                      margin={{ top: 5, right: 30, left: 20, bottom: 25 }}
                      syncId="monitoring"
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis 
                        dataKey="time" 
                        label={{ value: 'Time', position: 'insideBottomRight', offset: -10 }}
                        tick={{ fontSize: 12 }}
                        angle={-45}
                        textAnchor="end"
                        type="category"
                        allowDataOverflow={true}
                      />
                      <YAxis 
                        label={{ value: 'KB', angle: -90, position: 'insideLeft' }}
                        allowDataOverflow={true}
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
                        isAnimationActive={false}
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
                      ref={chartRef2}
                      data={chartData2}
                      margin={{ top: 5, right: 30, left: 20, bottom: 25 }}
                      syncId="monitoring"
                    >
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis 
                        dataKey="time" 
                        label={{ value: 'Time', position: 'insideBottomRight', offset: -10 }}
                        tick={{ fontSize: 12 }}
                        angle={-45}
                        textAnchor="end"
                        type="category"
                        allowDataOverflow={true}
                      />
                      <YAxis 
                        label={{ value: '%', angle: -90, position: 'insideLeft' }}
                        allowDataOverflow={true}
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
                        isAnimationActive={false}
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