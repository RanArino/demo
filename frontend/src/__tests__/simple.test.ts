/**
 * Simple test to verify Jest setup is working
 */

describe('Jest Setup', () => {
  it('should run basic test', () => {
    expect(1 + 1).toBe(2);
  });

  it('should have test environment configured', () => {
    expect(process.env.NODE_ENV).toBe('test');
  });
});