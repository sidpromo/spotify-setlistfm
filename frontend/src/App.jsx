import { useState, useEffect } from 'react';
import { isLoggedIn, setTokens, clearTokens } from './lib/api';
import Login from './components/Login';
import Search from './components/Search';
import PlaylistResult from './components/PlaylistResult';
import Callback from './components/Callback';

function App() {
  const [loggedIn, setLoggedIn] = useState(isLoggedIn());
  const [result, setResult] = useState(null);
  const [isCallback, setIsCallback] = useState(false);

  // Check if this is the OAuth callback page
  useEffect(() => {
    const path = window.location.pathname;
    if (path === '/callback') {
      setIsCallback(true);
    }
  }, []);

  function handleLoginComplete(data) {
    setTokens(data.accessToken, data.refreshToken);
    localStorage.setItem('userId', data.userId);
    localStorage.setItem('displayName', data.displayName);
    setLoggedIn(true);
    setIsCallback(false);
    window.history.replaceState({}, '', '/');
  }

  function handleLogout() {
    clearTokens();
    setLoggedIn(false);
    setResult(null);
  }

  function handlePlaylistCreated(jobResult) {
    setResult(jobResult);
  }

  // OAuth callback handling
  if (isCallback) {
    return <Callback onComplete={handleLoginComplete} />;
  }

  if (!loggedIn) {
    return <Login />;
  }

  return (
    <div className="min-h-screen bg-gray-950 text-white">
      <header className="border-b border-gray-800 px-6 py-4 flex items-center justify-between">
        <h1 className="text-xl font-bold text-green-400">🎵 Setlist Spotify</h1>
        <div className="flex items-center gap-4">
          <span className="text-sm text-gray-400">
            {localStorage.getItem('displayName') || 'User'}
          </span>
          <button
            onClick={handleLogout}
            className="text-sm text-gray-400 hover:text-white transition"
          >
            Logout
          </button>
        </div>
      </header>

      <main className="max-w-2xl mx-auto px-6 py-12">
        {result ? (
          <PlaylistResult result={result} onBack={() => setResult(null)} />
        ) : (
          <Search onPlaylistCreated={handlePlaylistCreated} />
        )}
      </main>
    </div>
  );
}

export default App;
