import axios from 'axios';

const apiClient = axios.create({
  baseURL: process.env.REACT_APP_IDENTITY_API_URL || 'http://localhost:8085/api/v1/identity',
});

apiClient.interceptors.request.use(
  config => {
    const token = localStorage.getItem('jwtToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  error => {
    return Promise.reject(error);
  }
);

export default apiClient;
