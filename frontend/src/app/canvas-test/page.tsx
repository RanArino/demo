'use client';

import { useState } from 'react';

export default function CanvasTestPage() {
  const [nodeIds, setNodeIds] = useState('');
  const [searchQuery, setSearchQuery] = useState('');
  const [spaceId, setSpaceId] = useState('');
  const [topK, setTopK] = useState('10');
  const [filterSpaceId, setFilterSpaceId] = useState('');
  const [limit, setLimit] = useState('25');
  const [results, setResults] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const clearResults = () => {
    setResults(null);
    setError(null);
  };

  const fetchNodes = async () => {
    setLoading(true);
    setError(null);
    try {
      const { getNodes } = await import('../../api/actions/canvasActions');
      const idsArray = nodeIds.split(',').map(id => id.trim()).filter(Boolean);

      if (idsArray.length === 0) {
        throw new Error('Please enter at least one node ID');
      }

      const result = await getNodes(idsArray);
      setResults(result);

      if (!result.success) {
        setError((result.error as any)?.message || 'Failed to fetch nodes');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch nodes';
      setError(errorMessage);
      setResults({ success: false, error: errorMessage });
    }
    setLoading(false);
  };

  const performSemanticSearch = async () => {
    setLoading(true);
    setError(null);
    try {
      const { semanticSearch } = await import('../../api/actions/canvasActions');

      if (!searchQuery.trim()) {
        throw new Error('Please enter a search query');
      }

      if (!spaceId.trim()) {
        throw new Error('Please enter a space ID');
      }

      const result = await semanticSearch({
        query: searchQuery.trim(),
        spaceId: spaceId.trim(),
        topK: parseInt(topK) || 10,
        nodeTypes: [], // Can be extended to allow user input
      });

      setResults(result);

      if (!result.success) {
        setError((result.error as any)?.message || 'Failed to perform semantic search');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to perform semantic search';
      setError(errorMessage);
      setResults({ success: false, error: errorMessage });
    }
    setLoading(false);
  };

  const performSearchNodes = async () => {
    setLoading(true);
    setError(null);
    try {
      const { searchNodes } = await import('../../api/actions/canvasActions');

      const result = await searchNodes({
        filter: filterSpaceId.trim() ? {
          spaceId: filterSpaceId.trim(),
        } : undefined,
        limit: parseInt(limit) || 25,
      });

      setResults(result);

      if (!result.success) {
        setError((result.error as any)?.message || 'Failed to search nodes');
      }
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to search nodes';
      setError(errorMessage);
      setResults({ success: false, error: errorMessage });
    }
    setLoading(false);
  };

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <h1 className="text-3xl font-bold mb-6">Canvas Service Test</h1>

      {/* Get Nodes Section */}
      <div className="mb-8 p-4 border rounded-lg">
        <h2 className="text-xl font-semibold mb-3">Get Nodes by IDs</h2>
        <div className="flex flex-col gap-2">
          <input
            type="text"
            placeholder="Enter node IDs (comma-separated)"
            value={nodeIds}
            onChange={(e) => setNodeIds(e.target.value)}
            className="border rounded px-3 py-2"
            disabled={loading}
          />
          <button
            onClick={fetchNodes}
            disabled={loading}
            className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded disabled:opacity-50"
          >
            {loading ? 'Loading...' : 'Fetch Nodes'}
          </button>
        </div>
      </div>

      {/* Semantic Search Section */}
      <div className="mb-8 p-4 border rounded-lg">
        <h2 className="text-xl font-semibold mb-3">Semantic Search</h2>
        <div className="flex flex-col gap-2">
          <input
            type="text"
            placeholder="Search query"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="border rounded px-3 py-2"
            disabled={loading}
          />
          <input
            type="text"
            placeholder="Space ID"
            value={spaceId}
            onChange={(e) => setSpaceId(e.target.value)}
            className="border rounded px-3 py-2"
            disabled={loading}
          />
          <input
            type="number"
            placeholder="Top K results"
            value={topK}
            onChange={(e) => setTopK(e.target.value)}
            className="border rounded px-3 py-2"
            disabled={loading}
            min="1"
            max="100"
          />
          <button
            onClick={performSemanticSearch}
            disabled={loading}
            className="bg-green-500 hover:bg-green-600 text-white px-4 py-2 rounded disabled:opacity-50"
          >
            {loading ? 'Loading...' : 'Search'}
          </button>
        </div>
      </div>

      {/* Search Nodes Section */}
      <div className="mb-8 p-4 border rounded-lg">
        <h2 className="text-xl font-semibold mb-3">Search Nodes (Filtered)</h2>
        <div className="flex flex-col gap-2">
          <input
            type="text"
            placeholder="Filter by Space ID (optional)"
            value={filterSpaceId}
            onChange={(e) => setFilterSpaceId(e.target.value)}
            className="border rounded px-3 py-2"
            disabled={loading}
          />
          <input
            type="number"
            placeholder="Limit"
            value={limit}
            onChange={(e) => setLimit(e.target.value)}
            className="border rounded px-3 py-2"
            disabled={loading}
            min="1"
            max="1000"
          />
          <button
            onClick={performSearchNodes}
            disabled={loading}
            className="bg-purple-500 hover:bg-purple-600 text-white px-4 py-2 rounded disabled:opacity-50"
          >
            {loading ? 'Loading...' : 'Search Nodes'}
          </button>
        </div>
      </div>

      {/* Clear Results Button */}
      {(results || error) && (
        <button
          onClick={clearResults}
          className="mb-4 bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded"
        >
          Clear Results
        </button>
      )}

      {/* Error Display */}
      {error && (
        <div className="mb-4 p-4 bg-red-50 border border-red-200 rounded-lg">
          <h3 className="text-red-800 font-semibold mb-2">Error</h3>
          <pre className="text-red-700 text-sm whitespace-pre-wrap">{error}</pre>
        </div>
      )}

      {/* Results Display */}
      {results && (
        <div className="p-4 border rounded-lg">
          <div className="flex items-center justify-between mb-2">
            <h2 className="text-lg font-semibold">Results</h2>
            <span className={`px-2 py-1 rounded text-sm ${
              results.success
                ? 'bg-green-100 text-green-800'
                : 'bg-red-100 text-red-800'
            }`}>
              {results.success ? 'Success' : 'Error'}
            </span>
          </div>
          <div className="bg-gray-50 p-4 rounded overflow-auto max-h-96">
            <pre className="text-sm whitespace-pre-wrap">
              {JSON.stringify(results, null, 2)}
            </pre>
          </div>

          {/* Quick Stats for Results */}
          {results.success && results.data && (
            <div className="mt-4 p-3 bg-blue-50 rounded">
              <h3 className="font-semibold text-blue-800 mb-2">Quick Stats</h3>
              <div className="text-blue-700 text-sm">
                {results.data.nodes && (
                  <p>• Nodes returned: {results.data.nodes.length}</p>
                )}
                {results.data.results && (
                  <p>• Search results: {results.data.results.length}</p>
                )}
                {results.data.results?.[0]?.score && (
                  <p>• Top score: {results.data.results[0].score.toFixed(4)}</p>
                )}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}