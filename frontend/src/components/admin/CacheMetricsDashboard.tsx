'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { RefreshCw, Download, Trash2, AlertTriangle, CheckCircle, Clock } from 'lucide-react';

interface CacheMetrics {
  hits: number;
  misses: number;
  errors: number;
  totalRequests: number;
  avgResponseTime: number;
  lastUpdated: number;
}

interface OperationMetrics extends CacheMetrics {
  name: string;
  hitRate: number;
}

interface MetricsResponse {
  timestamp: string;
  summary: {
    totalOperations: number;
    totalRequests: number;
    totalHits: number;
    totalMisses: number;
    totalErrors: number;
    overallHitRate: number;
    averageResponseTime: number;
  };
  operations: OperationMetrics[];
}

export default function CacheMetricsDashboard() {
  const [metrics, setMetrics] = useState<MetricsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(true);

  const fetchMetrics = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/cache-metrics');
      
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }
      
      const data = await response.json();
      setMetrics(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch metrics');
    } finally {
      setLoading(false);
    }
  };

  const resetMetrics = async (operation?: string) => {
    try {
      const response = await fetch('/api/cache-metrics', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action: 'reset', operation }),
      });
      
      if (response.ok) {
        await fetchMetrics();
      }
    } catch (err) {
      console.error('Failed to reset metrics:', err);
    }
  };

  const exportMetrics = async () => {
    try {
      const response = await fetch('/api/cache-metrics?format=export');
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `cache-metrics-${new Date().toISOString()}.json`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
    } catch (err) {
      console.error('Failed to export metrics:', err);
    }
  };

  useEffect(() => {
    fetchMetrics();
  }, []);

  useEffect(() => {
    if (!autoRefresh) return;
    
    const interval = setInterval(fetchMetrics, 30000); // Refresh every 30 seconds
    return () => clearInterval(interval);
  }, [autoRefresh]);

  const getStatusColor = (hitRate: number, errorRate: number): "default" | "destructive" | "secondary" | "outline" => {
    if (errorRate > 10) return 'destructive';
    if (hitRate < 50) return 'outline'; // Changed from 'warning' to 'outline'
    if (hitRate > 80) return 'default'; // Changed from 'success' to 'default'
    return 'secondary';
  };

  const getStatusIcon = (hitRate: number, errorRate: number) => {
    if (errorRate > 10) return <AlertTriangle className="h-4 w-4" />;
    if (hitRate > 80) return <CheckCircle className="h-4 w-4" />;
    return <Clock className="h-4 w-4" />;
  };

  const formatTime = (timestamp: number) => {
    return new Date(timestamp).toLocaleString();
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms.toFixed(2)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  if (loading && !metrics) {
    return (
      <div className="p-6">
        <div className="flex items-center justify-center h-32">
          <RefreshCw className="h-6 w-6 animate-spin" />
          <span className="ml-2">Loading cache metrics...</span>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <Card className="border-destructive">
          <CardContent className="pt-6">
            <div className="flex items-center">
              <AlertTriangle className="h-5 w-5 text-destructive mr-2" />
              <span className="text-destructive">Error loading metrics: {error}</span>
            </div>
            <Button onClick={fetchMetrics} className="mt-4" variant="outline" size="sm">
              <RefreshCw className="h-4 w-4 mr-2" />
              Retry
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  if (!metrics) {
    return (
      <div className="p-6">
        <Card>
          <CardContent className="pt-6">
            <p className="text-muted-foreground">No cache metrics available</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Cache Metrics Dashboard</h1>
          <p className="text-muted-foreground">
            Monitor cache performance and hit rates across all operations
          </p>
        </div>
        
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => setAutoRefresh(!autoRefresh)}
          >
            <RefreshCw className={`h-4 w-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
            {autoRefresh ? 'Auto' : 'Manual'}
          </Button>
          
          <Button variant="outline" size="sm" onClick={fetchMetrics} disabled={loading}>
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
          
          <Button variant="outline" size="sm" onClick={exportMetrics}>
            <Download className="h-4 w-4 mr-2" />
            Export
          </Button>
          
          <Button 
            variant="outline" 
            size="sm" 
            onClick={() => resetMetrics()}
            className="text-destructive hover:text-destructive"
          >
            <Trash2 className="h-4 w-4 mr-2" />
            Reset All
          </Button>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Operations</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.summary.totalOperations}</div>
            <p className="text-xs text-muted-foreground">
              Cached operations being monitored
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Overall Hit Rate</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {metrics.summary.overallHitRate.toFixed(1)}%
            </div>
            <p className="text-xs text-muted-foreground">
              {metrics.summary.totalHits} hits / {metrics.summary.totalRequests} requests
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Average Response Time</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {formatDuration(metrics.summary.averageResponseTime)}
            </div>
            <p className="text-xs text-muted-foreground">
              Across all cached operations
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Errors</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics.summary.totalErrors}</div>
            <p className="text-xs text-muted-foreground">
              Error rate: {metrics.summary.totalRequests > 0 
                ? ((metrics.summary.totalErrors / metrics.summary.totalRequests) * 100).toFixed(1)
                : 0}%
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Operations Table */}
      <Card>
        <CardHeader>
          <CardTitle>Operation Details</CardTitle>
          <CardDescription>
            Performance metrics for individual cache operations
          </CardDescription>
        </CardHeader>
        <CardContent>
          {metrics.operations.length === 0 ? (
            <p className="text-muted-foreground text-center py-8">
              No operations have been recorded yet
            </p>
          ) : (
            <div className="space-y-4">
              {metrics.operations.map((operation) => {
                const errorRate = operation.totalRequests > 0 
                  ? (operation.errors / operation.totalRequests) * 100 
                  : 0;
                
                return (
                  <div 
                    key={operation.name}
                    className="border rounded-lg p-4 space-y-3"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        {getStatusIcon(operation.hitRate, errorRate)}
                        <h3 className="font-semibold">{operation.name}</h3>
                        <Badge variant={getStatusColor(operation.hitRate, errorRate)}>
                          {operation.hitRate.toFixed(1)}% hit rate
                        </Badge>
                      </div>
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => resetMetrics(operation.name)}
                        className="text-muted-foreground hover:text-destructive"
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                    
                    <div className="grid grid-cols-2 md:grid-cols-5 gap-4 text-sm">
                      <div>
                        <div className="text-muted-foreground">Requests</div>
                        <div className="font-medium">{operation.totalRequests}</div>
                      </div>
                      
                      <div>
                        <div className="text-muted-foreground">Hits</div>
                        <div className="font-medium text-green-600">{operation.hits}</div>
                      </div>
                      
                      <div>
                        <div className="text-muted-foreground">Misses</div>
                        <div className="font-medium text-yellow-600">{operation.misses}</div>
                      </div>
                      
                      <div>
                        <div className="text-muted-foreground">Errors</div>
                        <div className="font-medium text-red-600">{operation.errors}</div>
                      </div>
                      
                      <div>
                        <div className="text-muted-foreground">Avg Response</div>
                        <div className="font-medium">{formatDuration(operation.avgResponseTime)}</div>
                      </div>
                    </div>
                    
                    <div className="text-xs text-muted-foreground">
                      Last updated: {formatTime(operation.lastUpdated)}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Footer */}
      <div className="text-xs text-muted-foreground text-center">
        Last refreshed: {formatTime(new Date(metrics.timestamp).getTime())}
      </div>
    </div>
  );
}