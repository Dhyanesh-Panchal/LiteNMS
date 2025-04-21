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
  Button,
  TextField,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  Alert,
} from '@mui/material';
import { Add as AddIcon, Edit as EditIcon, Delete as DeleteIcon } from '@mui/icons-material';
import { credentialProfileService } from '../services/api';

function CredentialProfiles() {
  const [profiles, setProfiles] = useState([]);
  const [open, setOpen] = useState(false);
  const [error, setError] = useState(null);
  const [editingProfile, setEditingProfile] = useState(null);
  const [formData, setFormData] = useState({
    hostname: '',
    password: '',
    port: 22,
  });

  useEffect(() => {
    console.log('Component mounted, loading profiles...');
    loadProfiles();
  }, []);

  const loadProfiles = async () => {
    try {
      console.log('Fetching profiles...');
      const response = await credentialProfileService.getAll();
      console.log('Received response:', response.data);
      setProfiles(response.data.profiles || []);
      setError(null);
    } catch (error) {
      console.error('Error loading profiles:', error);
      setError('Failed to load credential profiles. Please try again later.');
    }
  };

  const handleOpen = () => {
    setOpen(true);
  };

  const handleClose = () => {
    setOpen(false);
    setEditingProfile(null);
    setFormData({ hostname: '', password: '', port: 22 });
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

  const handleDelete = async (profileId) => {
    try {
      console.log('Deleting profile with ID:', profileId);
      await credentialProfileService.delete(profileId);
      loadProfiles();
    } catch (error) {
      console.error('Error deleting profile:', error);
      setError('Failed to delete credential profile. Please try again.');
    }
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    try {
      const submitData = {
        ...formData,
        port: parseInt(formData.port, 10)
      };
      
      if (editingProfile) {
        console.log('Updating profile:', submitData);
        await credentialProfileService.update(editingProfile.id, submitData);
      } else {
        console.log('Creating profile:', submitData);
        await credentialProfileService.create(submitData);
      }
      handleClose();
      loadProfiles();
    } catch (error) {
      console.error('Error saving profile:', error);
      setError(`Failed to ${editingProfile ? 'update' : 'create'} credential profile. Please try again.`);
    }
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
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      <Dialog open={open} onClose={handleClose}>
        <DialogTitle>
          {editingProfile ? 'Edit Credential Profile' : 'Add Credential Profile'}
        </DialogTitle>
        <form onSubmit={handleSubmit}>
          <DialogContent>
            <TextField
              autoFocus
              margin="dense"
              name="hostname"
              label="Hostname"
              type="text"
              fullWidth
              value={formData.hostname}
              onChange={handleChange}
              required
            />
            <TextField
              margin="dense"
              name="password"
              label="Password"
              type="password"
              fullWidth
              value={formData.password}
              onChange={handleChange}
              required
            />
            <TextField
              margin="dense"
              name="port"
              label="Port"
              type="number"
              fullWidth
              value={formData.port}
              onChange={handleChange}
              required
            />
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
}

export default CredentialProfiles; 