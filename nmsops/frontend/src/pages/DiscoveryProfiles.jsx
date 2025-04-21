import React, { useState, useEffect } from 'react';
import {
  Box,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Typography,
  IconButton,
  Chip,
  Alert,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  ListItemText,
  Checkbox,
  OutlinedInput,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const ITEM_HEIGHT = 48;
const ITEM_PADDING_TOP = 8;
const MenuProps = {
  PaperProps: {
    style: {
      maxHeight: ITEM_HEIGHT * 4.5 + ITEM_PADDING_TOP,
      width: 250,
    },
  },
};

const DiscoveryProfiles = () => {
  const [profiles, setProfiles] = useState([]);
  const [credentialProfiles, setCredentialProfiles] = useState([]);
  const [open, setOpen] = useState(false);
  const [error, setError] = useState('');
  const [editingProfile, setEditingProfile] = useState(null);
  const [formData, setFormData] = useState({
    device_ips: '',
    credential_profile_ids: [],
  });
  const [ipError, setIpError] = useState('');

  useEffect(() => {
    loadProfiles();
    loadCredentialProfiles();
  }, []);

  const loadCredentialProfiles = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/credential-profiles`);
      setCredentialProfiles(response.data.profiles || []);
    } catch (err) {
      setError('Failed to load credential profiles');
    }
  };

  const loadProfiles = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/discovery-profiles`);
      console.log('Loaded profiles:', response.data.profiles); // Debug log
      setProfiles(response.data.profiles || []);
    } catch (err) {
      setError('Failed to load discovery profiles');
    }
  };

  const handleOpen = () => {
    setOpen(true);
    setError('');
  };

  const handleClose = () => {
    setOpen(false);
    setEditingProfile(null);
    setFormData({
      device_ips: '',
      credential_profile_ids: [],
    });
  };

  const handleEdit = (profile) => {
    setEditingProfile(profile);
    setFormData({
      device_ips: profile.device_ips.map(ip => formatIP(ip)).join(', '),
      credential_profile_ids: profile.credential_profile_ids,
    });
    setOpen(true);
  };

  const handleDelete = async (id) => {
    console.log('Attempting to delete profile with ID:', id);
    if (!id) {
      setError('Invalid profile ID');
      return;
    }

    if (window.confirm('Are you sure you want to delete this profile?')) {
      try {
        // Try a POST request with a method param instead of DELETE
        // Some backends, especially simple ones, might not implement DELETE directly
        const response = await axios.post(`${API_BASE_URL}/discovery-profiles`, {
          method: 'delete',
          id: id
        });
        
        console.log('Delete response:', response);
        if (response.status === 200 || response.status === 204) {
          loadProfiles();
        } else {
          setError('Failed to delete profile: Unexpected response');
        }
      } catch (err) {
        console.error('Delete error details:', {
          status: err.response?.status,
          data: err.response?.data,
          profileId: id
        });
        
        // Try alternative endpoints if the first one failed
        try {
          // Try a different endpoint structure
          const response = await axios.delete(`${API_BASE_URL}/discovery-profile/${id}`);
          console.log('Delete response from alternate endpoint:', response);
          if (response.status === 200 || response.status === 204) {
            loadProfiles();
            return;
          }
        } catch (altErr) {
          console.error('Alternative delete attempt failed:', altErr.message);
        }
        
        // If we get here, all attempts failed
        if (err.response) {
          if (err.response.status === 404) {
            setError(`The backend does not support deleting profiles. Please contact your backend developer to implement this feature.`);
          } else {
            setError(`Failed to delete profile: ${err.response.data?.error || err.message}`);
          }
        } else {
          setError('Failed to delete profile: Network error');
        }
      }
    }
  };

  const validateIP = (ip) => {
    // Regular expression for IPv4 address
    const ipv4Regex = /^(\d{1,3}\.){3}\d{1,3}$/;
    
    if (!ipv4Regex.test(ip)) {
      return false;
    }

    // Check each octet is between 0 and 255
    const octets = ip.split('.');
    return octets.every(octet => {
      const num = parseInt(octet, 10);
      return num >= 0 && num <= 255;
    });
  };

  const handleChange = (event) => {
    const { name, value } = event.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));

    if (name === 'device_ips') {
      const ips = value.split(',').map(ip => ip.trim()).filter(ip => ip);
      const invalidIps = ips.filter(ip => !validateIP(ip));
      
      if (invalidIps.length > 0) {
        setIpError(`Invalid IP addresses: ${invalidIps.join(', ')}`);
      } else {
        setIpError('');
      }
    }
  };

  const handleCredentialProfileChange = (event) => {
    const { value } = event.target;
    setFormData(prev => ({
      ...prev,
      credential_profile_ids: value,
    }));
  };

  const ipToUint32 = (ip) => {
    const octets = ip.split('.').map(Number);
    return ((octets[0] << 24) | (octets[1] << 16) | (octets[2] << 8) | octets[3]) >>> 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    const ips = formData.device_ips.split(',').map(ip => ip.trim()).filter(ip => ip);
    const invalidIps = ips.filter(ip => !validateIP(ip));
    
    if (invalidIps.length > 0) {
      setError(`Invalid IP addresses: ${invalidIps.join(', ')}`);
      return;
    }

    try {
      const profileData = {
        device_ips: ips.map(ip => ipToUint32(ip)),
        credential_profile_ids: formData.credential_profile_ids,
      };

      if (editingProfile) {
        await axios.put(`${API_BASE_URL}/discovery-profiles/${editingProfile.id}`, profileData);
      } else {
        await axios.post(`${API_BASE_URL}/discovery-profiles`, profileData);
      }

      handleClose();
      loadProfiles();
    } catch (err) {
      setError(err.response?.data?.error || 'Failed to save profile');
    }
  };

  const formatIP = (ip) => {
    // Convert uint32 to dotted decimal
    return `${(ip >>> 24) & 255}.${(ip >>> 16) & 255}.${(ip >>> 8) & 255}.${ip & 255}`;
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
          Discovery Profiles
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={handleOpen}
        >
          Add Profile
        </Button>
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
              <TableCell>Device IPs</TableCell>
              <TableCell>Credential Profiles</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {profiles.length === 0 ? (
              <TableRow>
                <TableCell colSpan={3} align="center">
                  No discovery profiles found
                </TableCell>
              </TableRow>
            ) : (
              profiles.map((profile) => {
                console.log('Rendering profile:', profile);
                return (
                  <TableRow key={profile.id}>
                    <TableCell>
                      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                        {profile.device_ips.map((ip) => (
                          <Chip 
                            key={`${profile.id}-${ip}`} 
                            label={formatIP(ip)} 
                            size="small" 
                          />
                        ))}
                      </Box>
                    </TableCell>
                    <TableCell>
                      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                        {profile.credential_profile_ids.map((id) => {
                          const credentialProfile = credentialProfiles.find(cp => cp.id === id);
                          return credentialProfile ? (
                            <Chip 
                              key={`${profile.id}-${id}`}
                              label={`${credentialProfile.hostname}:${credentialProfile.port}`} 
                              size="small" 
                            />
                          ) : null;
                        })}
                      </Box>
                    </TableCell>
                    <TableCell>
                      <IconButton 
                        color="primary"
                        onClick={() => handleEdit(profile)}
                      >
                        <EditIcon />
                      </IconButton>
                      <IconButton 
                        color="error"
                        onClick={() => handleDelete(profile.id)}
                      >
                        <DeleteIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editingProfile ? 'Edit Discovery Profile' : 'Add Discovery Profile'}
        </DialogTitle>
        <form onSubmit={handleSubmit}>
          <DialogContent>
            <TextField
              autoFocus
              margin="dense"
              name="device_ips"
              label="Device IPs (comma-separated)"
              type="text"
              fullWidth
              value={formData.device_ips}
              onChange={handleChange}
              required
              error={!!ipError}
              helperText={ipError || "Enter IPs in dotted decimal format (e.g., 192.168.1.1, 10.0.0.1)"}
              sx={{ mb: 2 }}
            />
            <FormControl fullWidth>
              <InputLabel>Credential Profiles</InputLabel>
              <Select
                multiple
                value={formData.credential_profile_ids}
                onChange={handleCredentialProfileChange}
                input={<OutlinedInput label="Credential Profiles" />}
                renderValue={(selected) => (
                  <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                    {selected.map((id) => {
                      const profile = credentialProfiles.find(cp => cp.id === id);
                      return profile ? (
                        <Chip 
                          key={id} 
                          label={`${profile.hostname}:${profile.port}`} 
                          size="small"
                        />
                      ) : null;
                    })}
                  </Box>
                )}
                MenuProps={MenuProps}
              >
                {credentialProfiles.map((profile) => (
                  <MenuItem key={profile.id} value={profile.id}>
                    <Checkbox checked={formData.credential_profile_ids.indexOf(profile.id) > -1} />
                    <ListItemText 
                      primary={`${profile.hostname}:${profile.port}`}
                      secondary={`ID: ${profile.id}`}
                    />
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </DialogContent>
          <DialogActions>
            <Button onClick={handleClose}>Cancel</Button>
            <Button type="submit" variant="contained">
              {editingProfile ? 'Update' : 'Create'}
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    </Box>
  );
};

export default DiscoveryProfiles; 