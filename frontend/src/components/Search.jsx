import { useState, useEffect, useRef } from 'react';
import { searchArtists, createPlaylist, getJobStatus } from '../lib/api';

function Search({ onPlaylistCreated }) {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState([]);
  const [loading, setLoading] = useState(false);
  const [selectedArtist, setSelectedArtist] = useState(null);
  const [jobStatus, setJobStatus] = useState(null);
  const [error, setError] = useState(null);
  const debounceRef = useRef(null);

  // Debounced search
  useEffect(() => {
    if (query.length < 2) {
      setResults([]);
      return;
    }

    clearTimeout(debounceRef.current);
    debounceRef.current = setTimeout(async () => {
      setLoading(true);
      const data = await searchArtists(query);
      setResults(data.artists || []);
      setLoading(false);
    }, 300);

    return () => clearTimeout(debounceRef.current);
  }, [query]);

  async function handleCreatePlaylist(artist) {
    setSelectedArtist(artist);
    setError(null);
    setJobStatus({ status: 'pending' });

    const job = await createPlaylist(artist.mbid, artist.name);
    if (!job) {
      setError('Failed to create playlist. Please try again.');
      setJobStatus(null);
      return;
    }

    // Poll for job completion
    const pollInterval = setInterval(async () => {
      const status = await getJobStatus(job.jobId);
      if (!status) {
        clearInterval(pollInterval);
        setError('Failed to check job status.');
        setJobStatus(null);
        return;
      }

      setJobStatus(status);

      if (status.status === 'completed') {
        clearInterval(pollInterval);
        onPlaylistCreated(status.result);
      } else if (status.status === 'failed') {
        clearInterval(pollInterval);
        setError(status.error || 'Playlist creation failed.');
        setJobStatus(null);
      }
    }, 2000);
  }

  return (
    <div className="space-y-8">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-bold">Find an Artist</h2>
        <p className="text-gray-400">Search for a band and we'll predict their next setlist</p>
      </div>

      {/* Search Input */}
      <div className="relative">
        <input
          type="text"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Search artist (e.g. Metallica, ERRA, Spiritbox...)"
          className="w-full bg-gray-900 border border-gray-700 rounded-lg px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:border-green-500 transition"
          disabled={!!jobStatus}
        />
        {loading && (
          <div className="absolute right-3 top-3.5">
            <div className="w-5 h-5 border-2 border-gray-600 border-t-green-400 rounded-full animate-spin"></div>
          </div>
        )}
      </div>

      {/* Search Results */}
      {results.length > 0 && !jobStatus && (
        <div className="bg-gray-900 border border-gray-800 rounded-lg overflow-hidden">
          {results.slice(0, 8).map((artist) => (
            <button
              key={artist.mbid}
              onClick={() => handleCreatePlaylist(artist)}
              className="w-full text-left px-4 py-3 hover:bg-gray-800 transition border-b border-gray-800 last:border-b-0"
            >
              <div className="font-medium">{artist.name}</div>
              {artist.disambiguation && (
                <div className="text-sm text-gray-500">{artist.disambiguation}</div>
              )}
            </button>
          ))}
        </div>
      )}

      {/* Job Progress */}
      {jobStatus && (
        <div className="bg-gray-900 border border-gray-800 rounded-lg p-6 text-center space-y-4">
          <div className="w-8 h-8 border-2 border-gray-600 border-t-green-400 rounded-full animate-spin mx-auto"></div>
          <div>
            <p className="font-medium">Creating playlist for {selectedArtist?.name}...</p>
            <p className="text-sm text-gray-400 mt-1">
              {jobStatus.status === 'pending' && 'Queued...'}
              {jobStatus.status === 'processing' && 'Fetching setlists and creating playlist...'}
            </p>
          </div>
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="bg-red-900/20 border border-red-800 rounded-lg p-4 text-center">
          <p className="text-red-400">{error}</p>
          <button
            onClick={() => { setError(null); setSelectedArtist(null); }}
            className="mt-2 text-sm text-gray-400 hover:text-white"
          >
            Try again
          </button>
        </div>
      )}
    </div>
  );
}

export default Search;
