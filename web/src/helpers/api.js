import { getUserIdFromLocalStorage, showError } from './utils';
import axios from 'axios';

function createAPIInstance() {
  const instance = axios.create({
    baseURL: import.meta.env.VITE_REACT_APP_SERVER_URL
      ? import.meta.env.VITE_REACT_APP_SERVER_URL
      : '',
    headers: {
      'New-Api-User': getUserIdFromLocalStorage(),
      'Cache-Control': 'no-store',
    },
  });

  instance.interceptors.request.use((config) => {
    config.headers['New-Api-User'] = getUserIdFromLocalStorage();
    return config;
  });

  instance.interceptors.response.use(
    (response) => response,
    (error) => {
      showError(error);
    },
  );

  return instance;
}

export let API = createAPIInstance();

export function updateAPI() {
  API = createAPIInstance();
}