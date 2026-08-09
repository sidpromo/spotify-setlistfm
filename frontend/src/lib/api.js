const API_URL = import.meta.env.VITE_API_URL || '';

function getToken() {
  return localStorage.getItem('accessToken');
}

export function setTokens(accessToken, refreshToken) {
  localStorage.setItem('accessToken', accessToken);
  localStorage.setItem('refreshToken', refreshToken);
}

export function clearTokens() {
  localStorage.removeItem('accessToken');
  localStorage.removeItem('refreshToken');
  localStorage.removeItem('userId');
  localStorage.removeItem('displayName');
}

export function isLoggedIn() {
  return !!getToken();
}

async function request(path, options = {}) {
  const token = getToken();
  const headers = { ...options.headers };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  if (options.body) {
    headers['Content-Type'] = 'application/json';
  }

  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  });

  if (res.status === 401) {
    clearTokens();
    window.location.reload();
    return null;
  }

  return res;
}

export async function searchArtists(query) {
  const res = await request(`/v1/artists/search?q=${encodeURIComponent(query)}`);
  if (!res || !res.ok) return { artists: [] };
  return res.json();
}

export async function getSetlists(mbid) {
  const res = await request(`/v1/artists/${mbid}/setlists`);
  if (!res || !res.ok) return null;
  return res.json();
}

export async function createPlaylist(artistMbid, artistName) {
  const res = await request('/v1/playlists', {
    method: 'POST',
    body: JSON.stringify({ artistMbid, artistName }),
  });
  if (!res || !res.ok) return null;
  return res.json();
}

export async function getJobStatus(jobId) {
  const res = await request(`/v1/playlists/jobs/${jobId}`);
  if (!res || !res.ok) return null;
  return res.json();
}

export function getSpotifyLoginUrl() {
  return `${API_URL}/v1/auth/spotify/login`;
}
