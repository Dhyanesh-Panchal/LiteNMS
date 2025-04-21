import { useState, useEffect } from 'react';
import {
  Box,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Switch,
  IconButton,
  Alert,
} from '@mui/material';
import { Refresh as RefreshIcon } from '@mui/icons-material';
import { deviceService } from '../services/api';

function Devices() {
  const [devices, setDevices] = useState([]);
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadDevices();
  }, []);

  const loadDevices = async () => {
    try {
      setLoading(true);
      const response = await deviceService.getAll();
      setDevices(response.data.devices || []);
      setError(null);
    } catch (error) {
      console.error('Error loading devices:', error);
      setError('Failed to load devices. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  const handleProvisionToggle = async (deviceIP) => {
    try {
      const device = devices.find(d => d.ip === deviceIP);
      if (!device) return;

      await deviceService.updateProvisionStatus(deviceIP, !device.is_provisioned);
      loadDevices();
    } catch (error) {
      console.error('Error updating provision status:', error);
      setError('Failed to update device provision status. Please try again.');
    }
  };

  const formatIP = (ip) => {
    // Convert uint32 to dotted decimal notation
    return `${(ip >> 24) & 0xFF}.${(ip >> 16) & 0xFF}.${(ip >> 8) & 0xFF}.${ip & 0xFF}`;
  };

  return (
    <Box sx={{ 
      width: '100%',
      maxWidth: 1200,
      margin: '0 auto',
      p: 3,
    }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">
          Devices
        </Typography>
        <IconButton 
          onClick={loadDevices}
          disabled={loading}
        >
          <RefreshIcon />
        </IconButton>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>IP Address</TableCell>
              <TableCell>Credential Profile ID</TableCell>
              <TableCell>Provisioned</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {devices.length === 0 ? (
              <TableRow>
                <TableCell colSpan={4} align="center">
                  No devices found
                </TableCell>
              </TableRow>
            ) : (
              devices.map((device) => (
                <TableRow key={device.ip}>
                  <TableCell>{formatIP(device.ip)}</TableCell>
                  <TableCell>{device.credential_id || 'None'}</TableCell>
                  <TableCell>
                    <Switch
                      checked={device.is_provisioned}
                      onChange={() => handleProvisionToggle(device.ip)}
                    />
                  </TableCell>
                  <TableCell>
                    {/* We'll add more actions here later */}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}

export default Devices; 