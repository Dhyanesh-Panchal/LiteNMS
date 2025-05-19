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
  Alert,
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import VisibilityIcon from '@mui/icons-material/Visibility';
import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const CredentialProfiles = () => {
  const [profiles, setProfiles] = useState([]);
  const [open, setOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [selectedProfile, setSelectedProfile] = useState(null);
  const [error, setError] = useState('');
  const [editingProfile, setEditingProfile] = useState(null);
  const [formData, setFormData] = useState({
    hostname: '',
    password: '',
    port: '',
  });

  useEffect(() => {
    loadProfiles();
  }, []);

  const loadProfiles = async () => {
    try {
      const response = await axios.get(`${API_BASE_URL}/credential-profiles`);
      setProfiles(response.data.profiles || []);
    } catch (err) {
      setError('Failed to load credential profiles');
    }
  };

  const handleOpen = () => {
    setOpen(true);
    setError('');
    setFormData({
      hostname: '',
      password: '',
      port: '',
    });
  };

  const handleClose = () => {
    setOpen(false);
    setEditingProfile(null);
    setFormData({
      hostname: '',
      password: '',
      port: '',
    });
  };

  const handleEdit = (profile) => {
    setEditingProfile(profile);
    setFormData({
      hostname: profile.hostname,
      password: profile.password,
      port: profile.port,
    });
    setOpen(true);
  };

  const handleChange = (event) => {
    const { name, value } = event.target;
    if(name == "port") {
      if(!Number.isInteger(Number(value))) {
        return;
      }

      if(value < 0 || value > 65535) {
        return;
      }

      setFormData(prev => ({
        ...prev,
        [name]: Number(value),
      }));
    }
    else {
      setFormData(prev => ({
        ...prev,
        [name]: value,
      }));
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    try {
      if (editingProfile) {
        await axios.put(`${API_BASE_URL}/credential-profiles/${editingProfile.id}`, formData);
      } else {
        await axios.post(`${API_BASE_URL}/credential-profiles`, formData);
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

  return (
    <Box sx={{ 
      width: '100%',
      maxWidth: 1200,
      margin: '0 auto',
      p: 3,
    }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
        <Typography variant="h4">
          Credential Profiles
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
              <TableCell>Hostname</TableCell>
              <TableCell>Port</TableCell>
              <TableCell>Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {profiles.length === 0 ? (
              <TableRow>
                <TableCell colSpan={3} align="center">
                  No credential profiles found
                </TableCell>
              </TableRow>
            ) : (
              profiles.map((profile) => (
                <TableRow key={profile.id}>
                  <TableCell>{profile.hostname}</TableCell>
                  <TableCell>{profile.port}</TableCell>
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
        maxWidth="sm" 
        fullWidth
      >
        <DialogTitle>
          Credential Profile Details
        </DialogTitle>
        <DialogContent>
          {selectedProfile && (
            <Box sx={{ mt: 2 }}>
              <Typography variant="h6" gutterBottom>
                Hostname
              </Typography>
              <Typography>{selectedProfile.hostname}</Typography>
              
              <Typography variant="h6" gutterBottom sx={{ mt: 2 }}>
                Port
              </Typography>
              <Typography>{selectedProfile.port}</Typography>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDetails}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* Create/Edit Dialog */}
      <Dialog 
        open={open} 
        onClose={handleClose} 
        maxWidth="sm" 
        fullWidth
      >
        <DialogTitle>
          {editingProfile ? 'Edit Credential Profile' : 'Create Credential Profile'}
        </DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2 }}>
            <TextField
              fullWidth
              label="Hostname"
              name="hostname"
              value={formData.hostname}
              onChange={handleChange}
              required
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Password"
              name="password"
              type="password"
              value={formData.password}
              onChange={handleChange}
              required
              sx={{ mb: 2 }}
            />
            <TextField
              fullWidth
              label="Port"
              name="port"
              value={formData.port}
              onChange={handleChange}
              required
              sx={{ mb: 2 }}
            />
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button 
            onClick={handleSubmit} 
            variant="contained" 
            color="primary"
            disabled={!formData.hostname || !formData.password || !formData.port}
          >
            {editingProfile ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default CredentialProfiles; 