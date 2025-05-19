import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const credentialProfileService = {
  getAll: () => api.get('/credential-profiles'),
  create: (data) => api.post('/credential-profiles', data),
  update: (id, data) => api.put(`/credential-profiles/${id}`, data),
  delete: (id) => api.delete(`/credential-profiles/${id}`),
};

export const deviceService = {
  getAll: () => api.get('/devices'),
  updateProvisionStatus: (ips) => 
    api.put('/devices/update-provisioning', { provision_update_ips: ips }),
};

export const discoveryProfileService = {
  getAll: () => api.get('/discovery-profiles'),
  create: (data) => api.post('/discovery-profiles', data),
  update: (id, data) => api.put(`/discovery-profiles/${id}`, data),
  delete: (id) => api.delete(`/discovery-profiles/${id}`),
};

export default api; 