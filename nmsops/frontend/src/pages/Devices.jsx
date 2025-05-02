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
  Checkbox,
  Button,
  IconButton,
  Alert,
  Tooltip,
} from '@mui/material';
import { Refresh as RefreshIcon } from '@mui/icons-material';
import { deviceService } from '../services/api';

function Devices() {
  const [devices, setDevices] = useState([]);
  const [selectedDevices, setSelectedDevices] = useState(new Set());
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const [isUpdateMode, setIsUpdateMode] = useState(false);

  useEffect(() => {
    loadDevices();
  }, []);

  const loadDevices = async () => {
    try {
      setLoading(true);
      const response = await deviceService.getAll();
      // Sort devices by IP
      const sortedDevices = (response.data.devices || []).sort((a, b) => {
        const ipA = a.ip.split('.').map(Number);
        const ipB = b.ip.split('.').map(Number);
        for (let i = 0; i < 4; i++) {
          if (ipA[i] !== ipB[i]) {
            return ipA[i] - ipB[i];
          }
        }
        return 0;
      });
      setDevices(sortedDevices);
      setError(null);
    } catch (error) {
      console.error('Error loading devices:', error);
      setError('Failed to load devices. Please try again later.');
    } finally {
      setLoading(false);
    }
  };

  const handleCheckboxChange = (deviceIP) => {
    setSelectedDevices(prev => {
      const newSet = new Set(prev);
      if (newSet.has(deviceIP)) {
        newSet.delete(deviceIP);
      } else {
        newSet.add(deviceIP);
      }
      return newSet;
    });
  };

  const handleSelectAll = (event) => {
    if (event.target.checked) {
      const allIPs = new Set(devices.map(device => device.ip));
      setSelectedDevices(allIPs);
    } else {
      setSelectedDevices(new Set());
    }
  };

  const handleUpdateProvision = async () => {
    try {
      if (selectedDevices.size === 0) {
        setError('Please select at least one device');
        return;
      }

      const ipsToUpdate = Array.from(selectedDevices);
      await deviceService.updateProvisionStatus(ipsToUpdate);
      
      // Refresh the device list
      await loadDevices();
      
      // Reset update mode and selection
      setIsUpdateMode(false);
      setSelectedDevices(new Set());
    } catch (error) {
      console.error('Error updating provision status:', error);
      setError('Failed to update device provision status. Please try again.');
    }
  };

  const handleCancelUpdate = () => {
    setIsUpdateMode(false);
    setSelectedDevices(new Set());
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
        <Box>
          {!isUpdateMode ? (
            <Button 
              variant="contained" 
              color="primary" 
              onClick={() => setIsUpdateMode(true)}
              sx={{ mr: 2 }}
            >
              Update Provision
            </Button>
          ) : (
            <Box>
              <Button 
                variant="outlined" 
                color="error"
                onClick={handleCancelUpdate}
                sx={{ mr: 2 }}
              >
                Cancel
              </Button>
              <Button 
                variant="contained" 
                color="primary" 
                onClick={handleUpdateProvision}
                disabled={selectedDevices.size === 0}
              >
                Update Selected ({selectedDevices.size})
              </Button>
            </Box>
          )}
          <IconButton 
            onClick={loadDevices}
            disabled={loading}
          >
            <RefreshIcon />
          </IconButton>
        </Box>
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
              {isUpdateMode && (
                <TableCell>
                  <Tooltip title="Select/Deselect all devices">
                    <Checkbox
                      checked={selectedDevices.size === devices.length}
                      indeterminate={selectedDevices.size > 0 && selectedDevices.size < devices.length}
                      onChange={handleSelectAll}
                    />
                  </Tooltip>
                </TableCell>
              )}
              <TableCell>IP Address</TableCell>
              <TableCell>Credential Profile ID</TableCell>
              <TableCell>Provisioned</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {devices.length === 0 ? (
              <TableRow>
                <TableCell colSpan={isUpdateMode ? 4 : 3} align="center">
                  No devices found
                </TableCell>
              </TableRow>
            ) : (
              devices.map((device) => (
                <TableRow key={device.ip}>
                  {isUpdateMode && (
                    <TableCell>
                      <Checkbox
                        checked={selectedDevices.has(device.ip)}
                        onChange={() => handleCheckboxChange(device.ip)}
                      />
                    </TableCell>
                  )}
                  <TableCell>{device.ip}</TableCell>
                  <TableCell>{device.credential_id || 'None'}</TableCell>
                  <TableCell>
                    {device.is_provisioned ? 'Yes' : 'No'}
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