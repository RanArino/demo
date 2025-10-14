import { act, renderHook, waitFor } from '@testing-library/react';
import { useCanvasData } from '../useCanvasData';

const originalFetch = global.fetch;

describe('useCanvasData', () => {
  beforeEach(() => {
    jest.useFakeTimers();
    global.fetch = jest.fn();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    jest.useRealTimers();
    jest.clearAllMocks();
    global.fetch = originalFetch;
  });

  it('schedules a retry with backoff after a failed request and clears error on success', async () => {
    const firstError = new Error('Failed to load canvas data');
    const successPayload = { nodes: [] };

    (global.fetch as jest.Mock)
      .mockRejectedValueOnce(firstError)
      .mockResolvedValueOnce(
        {
          ok: true,
          json: async () => successPayload,
        },
      );

    const { result } = renderHook(() => useCanvasData('space-123'));

    await act(async () => {
      await Promise.resolve();
    });

    await waitFor(() => {
      expect(result.current.errorMessage).toEqual('Failed to load canvas data');
      expect(result.current.retryInMs).not.toBeNull();
    });

    act(() => {
      jest.advanceTimersByTime(2000);
    });

    await act(async () => {
      await Promise.resolve();
    });

    await waitFor(() => {
      expect(result.current.errorMessage).toBeNull();
      expect(result.current.retryInMs).toBeNull();
    });

    expect(global.fetch).toHaveBeenCalledTimes(2);
  });
});
