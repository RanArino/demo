/**
 * API endpoint for cache metrics monitoring
 * Provides cache performance data for monitoring and debugging
 */

import { NextRequest, NextResponse } from 'next/server';
import { auth } from '@clerk/nextjs/server';
import { 
  getAllCacheMetrics, 
  getCacheHitRate, 
  exportCacheMetricsForMonitoring,
  logCachePerformanceSummary,
  resetCacheMetrics
} from '@/api/actions/utils';

// Only allow access in development or with proper authentication
async function isAuthorized(): Promise<boolean> {
  try {
    if (process.env.NODE_ENV === 'development') {
      return true;
    }
    
    // In production, require authentication and admin privileges
    const { userId } = await auth();
    if (!userId) {
      return false;
    }
    
    // TODO: Add admin check here if needed
    // const user = await getUserById(userId);
    // return user?.role === 'admin';
    
    return true; // For now, allow authenticated users
  } catch {
    return false;
  }
}

export async function GET(request: NextRequest) {
  try {
    const authorized = await isAuthorized();
    if (!authorized) {
      return NextResponse.json(
        { error: 'Unauthorized' }, 
        { status: 401 }
      );
    }

    const { searchParams } = new URL(request.url);
    const format = searchParams.get('format') || 'json';
    const operation = searchParams.get('operation');

    if (format === 'export') {
      // Export format for external monitoring systems
      const exportData = exportCacheMetricsForMonitoring();
      return new NextResponse(exportData, {
        headers: {
          'Content-Type': 'application/json',
          'Content-Disposition': `attachment; filename="cache-metrics-${new Date().toISOString()}.json"`,
        },
      });
    }

    const allMetrics = getAllCacheMetrics();
    
    if (operation) {
      // Return metrics for specific operation
      const operationMetrics = allMetrics[operation];
      if (!operationMetrics) {
        return NextResponse.json(
          { error: `No metrics found for operation: ${operation}` },
          { status: 404 }
        );
      }
      
      return NextResponse.json({
        operation,
        metrics: operationMetrics,
        hitRate: getCacheHitRate(operation),
        timestamp: new Date().toISOString(),
      });
    }

    // Return all metrics with summary
    const operations = Object.keys(allMetrics);
    const summary = {
      totalOperations: operations.length,
      totalRequests: operations.reduce((sum, op) => sum + allMetrics[op].totalRequests, 0),
      totalHits: operations.reduce((sum, op) => sum + allMetrics[op].hits, 0),
      totalMisses: operations.reduce((sum, op) => sum + allMetrics[op].misses, 0),
      totalErrors: operations.reduce((sum, op) => sum + allMetrics[op].errors, 0),
      overallHitRate: 0,
      averageResponseTime: 0,
    };

    if (summary.totalRequests > 0) {
      summary.overallHitRate = (summary.totalHits / summary.totalRequests) * 100;
      summary.averageResponseTime = operations.reduce(
        (sum, op) => sum + allMetrics[op].avgResponseTime, 0
      ) / operations.length;
    }

    const response = {
      timestamp: new Date().toISOString(),
      summary,
      operations: Object.entries(allMetrics).map(([name, metrics]) => ({
        name,
        ...metrics,
        hitRate: getCacheHitRate(name),
      })),
    };

    return NextResponse.json(response);
  } catch (error) {
    console.error('Cache metrics API error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}

export async function POST(request: NextRequest) {
  try {
    const authorized = await isAuthorized();
    if (!authorized) {
      return NextResponse.json(
        { error: 'Unauthorized' }, 
        { status: 401 }
      );
    }

    const body = await request.json();
    const { action, operation } = body;

    switch (action) {
      case 'reset':
        resetCacheMetrics(operation);
        return NextResponse.json({ 
          message: operation 
            ? `Reset metrics for operation: ${operation}`
            : 'Reset all cache metrics'
        });

      case 'log-summary':
        logCachePerformanceSummary();
        return NextResponse.json({ 
          message: 'Cache performance summary logged to console'
        });

      default:
        return NextResponse.json(
          { error: `Unknown action: ${action}` },
          { status: 400 }
        );
    }
  } catch (error) {
    console.error('Cache metrics API error:', error);
    return NextResponse.json(
      { error: 'Internal server error' },
      { status: 500 }
    );
  }
}