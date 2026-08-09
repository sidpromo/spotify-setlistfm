function PlaylistResult({ result, onBack }) {
  return (
    <div className="space-y-8 text-center">
      <div className="space-y-2">
        <div className="text-5xl">🎉</div>
        <h2 className="text-2xl font-bold">Playlist Created!</h2>
        <p className="text-gray-400">{result.playlistName}</p>
      </div>

      <div className="bg-gray-900 border border-gray-800 rounded-lg p-6 space-y-4">
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div className="bg-gray-800 rounded-lg p-3">
            <div className="text-gray-400">Tracks Added</div>
            <div className="text-2xl font-bold text-green-400">{result.tracksAdded}</div>
          </div>
          <div className="bg-gray-800 rounded-lg p-3">
            <div className="text-gray-400">Based On</div>
            <div className="text-2xl font-bold text-green-400">{result.basedOnCount} shows</div>
          </div>
        </div>

        {result.tourName && (
          <p className="text-sm text-gray-400">
            Tour: <span className="text-white">{result.tourName}</span>
          </p>
        )}

        {result.notFound && result.notFound.length > 0 && (
          <div className="text-sm text-gray-500">
            <p>Couldn't find on Spotify: {result.notFound.join(', ')}</p>
          </div>
        )}
      </div>

      <div className="space-y-3">
        <a
          href={result.playlistUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-2 bg-green-500 hover:bg-green-400 text-black font-semibold px-8 py-3 rounded-full transition"
        >
          <svg className="w-5 h-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 0C5.4 0 0 5.4 0 12s5.4 12 12 12 12-5.4 12-12S18.66 0 12 0zm5.521 17.34c-.24.359-.66.48-1.021.24-2.82-1.74-6.36-2.101-10.561-1.141-.418.122-.779-.179-.899-.539-.12-.421.18-.78.54-.9 4.56-1.021 8.52-.6 11.64 1.32.42.18.479.659.301 1.02zm1.44-3.3c-.301.42-.841.6-1.262.3-3.239-1.98-8.159-2.58-11.939-1.38-.479.12-1.02-.12-1.14-.6-.12-.48.12-1.021.6-1.141C9.6 9.9 15 10.561 18.72 12.84c.361.181.54.78.241 1.2zm.12-3.36C15.24 8.4 8.82 8.16 5.16 9.301c-.6.179-1.2-.181-1.38-.721-.18-.601.18-1.2.72-1.381 4.26-1.26 11.28-1.02 15.721 1.621.539.3.719 1.02.419 1.56-.299.421-1.02.599-1.559.3z"/>
          </svg>
          Open in Spotify
        </a>

        <div>
          <button
            onClick={onBack}
            className="text-sm text-gray-400 hover:text-white transition"
          >
            ← Search another artist
          </button>
        </div>
      </div>
    </div>
  );
}

export default PlaylistResult;
