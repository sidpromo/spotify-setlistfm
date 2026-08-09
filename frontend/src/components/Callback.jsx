import { useEffect, useState } from 'react';

const API_URL = import.meta.env.VITE_API_URL || '';

function Callback({ onComplete }) {
  const [error, setError] = useState(null);

  useEffect(() => {
    // The backend OAuth callback returns JSON with tokens.
    // But since Spotify redirects the browser to the backend callback URL,
    // we need the backend to redirect to us with the tokens.
    // For now, we'll read from the page content if we're on the backend callback URL,
    // or handle a frontend callback route.

    // Check if there are tokens in the URL params (backend redirects here with tokens)
    const params = new URLSearchParams(window.location.search);
    const accessToken = params.get('accessToken');
    const refreshToken = params.get('refreshToken');
    const userId = params.get('userId');
    const displayName = params.get('displayName');

    if (accessToken && refreshToken) {
      onComplete({ accessToken, refreshToken, userId, displayName });
    } else {
      setError('Login failed. No token received.');
    }
  }, [onComplete]);

  if (error) {
    return (
      <div className="min-h-screen bg-gray-950 flex items-center justify-center">
        <div className="text-center space-y-4">
          <p className="text-red-400">{error}</p>
          <a href="/" className="text-sm text-gray-400 hover:text-white">
            ← Back to login
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-950 flex items-center justify-center">
      <div className="text-center space-y-4">
        <div className="w-8 h-8 border-2 border-gray-600 border-t-green-400 rounded-full animate-spin mx-auto"></div>
        <p className="text-gray-400">Logging in...</p>
      </div>
    </div>
  );
}

export default Callback;
