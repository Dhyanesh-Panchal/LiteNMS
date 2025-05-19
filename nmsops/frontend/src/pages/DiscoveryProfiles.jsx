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
  List,
  ListItem,
  Divider,
  ToggleButton,
  ToggleButtonGroup,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import VisibilityIcon from '@mui/icons-material/Visibility';
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
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedProfile, setSelectedProfile] = useState(null);
  const [error, setError] = useState('');
  const [editingProfile, setEditingProfile] = useState(null);
  const [formData, setFormData] = useState({
    device_ips: '',
    credential_profile_ids: [],
    is_cidr: false,
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
    setFormData({
      device_ips: '',
      credential_profile_ids: [],
      is_cidr: false,
    });
  };

  const handleClose = () => {
    setOpen(false);
    setEditingProfile(null);
    setFormData({
      device_ips: '',
      credential_profile_ids: [],
      is_cidr: false,
    });
  };

  const handleEdit = (profile) => {
    setEditingProfile(profile);
    setFormData({
      device_ips: profile.device_ips.join(', '),
      credential_profile_ids: profile.credential_profile_ids,
      is_cidr: false,
    });
    setOpen(true);
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

  const validateCIDR = (cidr) => {
    const cidrRegex = /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/;
    if (!cidrRegex.test(cidr)) {
      return false;
    }

    const [ip, mask] = cidr.split('/');
    if (!validateIP(ip)) {
      return false;
    }

    const maskNum = parseInt(mask, 10);
    return maskNum >= 0 && maskNum <= 32;
  };

  const handleChange = (event) => {
    const { name, value } = event.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));

    if (name === 'device_ips') {
      if (formData.is_cidr) {
        if (!validateCIDR(value)) {
          setIpError('Invalid CIDR notation');
        } else {
          setIpError('');
        }
      } else {
        const ips = value.split(',').map(ip => ip.trim()).filter(ip => ip);
        const invalidIps = ips.filter(ip => !validateIP(ip));
        
        if (invalidIps.length > 0) {
          setIpError(`Invalid IP addresses: ${invalidIps.join(', ')}`);
        } else {
          setIpError('');
        }
      }
    }
  };

  const handleInputTypeChange = (event, newInputType) => {
    if (newInputType !== null) {
      setFormData(prev => ({
        ...prev,
        is_cidr: newInputType === 'cidr',
        device_ips: '', // Clear the input when switching types
      }));
      setIpError('');
    }
  };

  const handleCredentialProfileChange = (event) => {
    const { value } = event.target;
    setFormData(prev => ({
      ...prev,
      credential_profile_ids: value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (formData.is_cidr) {
      if (!validateCIDR(formData.device_ips)) {
        setError('Invalid CIDR notation');
        return;
      }
    } else {
      const ips = formData.device_ips.split(',').map(ip => ip.trim()).filter(ip => ip);
      const invalidIps = ips.filter(ip => !validateIP(ip));
      
      if (invalidIps.length > 0) {
        setError(`Invalid IP addresses: ${invalidIps.join(', ')}`);
        return;
      }
    }

    try {
      const profileData = {
        device_ips: formData.is_cidr ? formData.device_ips : formData.device_ips.split(',').map(ip => ip.trim()).filter(ip => ip),
        credential_profile_ids: formData.credential_profile_ids,
        is_cidr: formData.is_cidr,
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

  const handleViewDetails = (profile) => {
    setSelectedProfile(profile);
    setDetailOpen(true);
  };

  const handleCloseDetails = () => {
    setDetailOpen(false);
    setSelectedProfile(null);
  };

  const renderLimitedIPs = (ips) => {
    const MAX_IPS = 5;
    if (ips.length <= MAX_IPS) {
      return ips.map((ip) => (
        <Chip
          key={ip}
          label={ip}
          sx={{ m: 0.5 }}
        />
      ));
    }
    return (
      <>
        {ips.slice(0, MAX_IPS).map((ip) => (
          <Chip
            key={ip}
            label={ip}
            sx={{ m: 0.5 }}
          />
        ))}
        <Chip
          label={`+${ips.length - MAX_IPS} more`}
          sx={{ m: 0.5 }}
        />
      </>
    );
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
              profiles.map((profile) => (
                <TableRow key={profile.id}>
                  <TableCell>
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 1 }}>
                      {renderLimitedIPs(profile.device_ips)}
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
                      onClick={() => handleViewDetails(profile)}
                    >
                      <VisibilityIcon />
                    </IconButton>
                    <IconButton 
                      color="primary"
                      onClick={() => handleEdit(profile)}
                    >
                      <EditIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Detail View Dialog */}
      <Dialog 
        open={detailOpen} 
        onClose={handleCloseDetails} 
        maxWidth="md" 
        fullWidth
      >
        <DialogTitle>
          Discovery Profile Details
        </DialogTitle>
        <DialogContent>
          {selectedProfile && (
            <Box sx={{ mt: 2 }}>
              <Typography variant="h6" gutterBottom>
                Device IPs
              </Typography>
              <List>
                {selectedProfile.device_ips.map((ip) => (
                  <ListItem key={ip}>
                    <Chip label={ip} />
                  </ListItem>
                ))}
              </List>
              
              <Divider sx={{ my: 2 }} />
              
              <Typography variant="h6" gutterBottom>
                Credential Profiles
              </Typography>
              <List>
                {selectedProfile.credential_profile_ids.map((id) => {
                  const credentialProfile = credentialProfiles.find(cp => cp.id === id);
                  return credentialProfile ? (
                    <ListItem key={id}>
                      <Chip 
                        label={`${credentialProfile.hostname}:${credentialProfile.port}`}
                        sx={{ mr: 1 }}
                      />
                      <Typography variant="body2" color="text.secondary">
                        ID: {credentialProfile.id}
                      </Typography>
                    </ListItem>
                  ) : null;
                })}
              </List>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDetails}>Close</Button>
        </DialogActions>
      </Dialog>

      <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editingProfile ? 'Edit Discovery Profile' : 'Create Discovery Profile'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2 }}>
            <ToggleButtonGroup
              value={formData.is_cidr ? 'cidr' : 'list'}
              exclusive
              onChange={handleInputTypeChange}
              aria-label="IP input type"
              sx={{ mb: 2 }}
            >
              <ToggleButton value="list" aria-label="list of IPs">
                List of IPs
              </ToggleButton>
              <ToggleButton value="cidr" aria-label="CIDR notation">
                CIDR Notation
              </ToggleButton>
            </ToggleButtonGroup>

            <TextField
              fullWidth
              label={formData.is_cidr ? "CIDR Notation (e.g., 192.168.1.0/24)" : "IP Addresses (comma-separated)"}
              name="device_ips"
              value={formData.device_ips}
              onChange={handleChange}
              error={!!ipError}
              helperText={ipError || (formData.is_cidr 
                ? "Enter CIDR notation (e.g., 192.168.1.0/24)" 
                : "Enter comma-separated IP addresses (e.g., 192.168.1.1, 192.168.1.2)")}
              sx={{ mb: 2 }}
            />

            <FormControl fullWidth sx={{ mb: 2 }}>
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
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button 
            onClick={handleSubmit} 
            variant="contained" 
            color="primary"
            disabled={!!ipError || formData.credential_profile_ids.length === 0}
          >
            {editingProfile ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default DiscoveryProfiles; 