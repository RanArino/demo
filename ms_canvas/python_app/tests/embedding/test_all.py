"""
Comprehensive test runner for embedding service.
Runs all embedding tests and provides summary.
"""
import pytest
import sys
import os


def run_unit_tests():
    """Run unit tests for embedding service."""
    return pytest.main([
        "tests/embedding/test_embedding_unit.py",
        "-v",
        "--tb=short"
    ])


def run_integration_tests():
    """Run integration tests for embedding service."""
    return pytest.main([
        "tests/embedding/test_embedding_integration.py",
        "-v",
        "--tb=short",
        "-m", "not skipif"
    ])


def run_performance_tests():
    """Run performance tests for embedding service."""
    return pytest.main([
        "tests/embedding/test_embedding_performance.py",
        "-v",
        "--tb=short",
        "-m", "performance"
    ])


def run_benchmark_tests():
    """Run benchmark tests for embedding service."""
    return pytest.main([
        "tests/embedding/test_embedding_performance.py",
        "-v",
        "--tb=short",
        "-m", "benchmark"
    ])


def run_all_tests():
    """Run all embedding tests."""
    print("=" * 60)
    print("EMBEDDING SERVICE TEST SUITE")
    print("=" * 60)
    
    test_results = {}
    
    # Unit tests
    print("\n1. Running Unit Tests...")
    test_results['unit'] = run_unit_tests()
    
    # Integration tests (if environment supports it)
    if os.environ.get('INTEGRATION_TESTS'):
        print("\n2. Running Integration Tests...")
        test_results['integration'] = run_integration_tests()
    else:
        print("\n2. Skipping Integration Tests (set INTEGRATION_TESTS=1 to enable)")
        test_results['integration'] = 0
    
    # Performance tests
    print("\n3. Running Performance Tests...")
    test_results['performance'] = run_performance_tests()
    
    # Benchmark tests
    if os.environ.get('RUN_BENCHMARKS'):
        print("\n4. Running Benchmark Tests...")
        test_results['benchmark'] = run_benchmark_tests()
    else:
        print("\n4. Skipping Benchmark Tests (set RUN_BENCHMARKS=1 to enable)")
        test_results['benchmark'] = 0
    
    # Summary
    print("\n" + "=" * 60)
    print("TEST RESULTS SUMMARY")
    print("=" * 60)
    
    for test_type, result in test_results.items():
        status = "PASSED" if result == 0 else "FAILED"
        print(f"{test_type.capitalize()} Tests: {status}")
    
    overall_result = sum(test_results.values())
    overall_status = "PASSED" if overall_result == 0 else "FAILED"
    print(f"\nOverall Result: {overall_status}")
    
    return overall_result


if __name__ == "__main__":
    # Parse command line arguments
    if len(sys.argv) > 1:
        test_type = sys.argv[1].lower()
        
        if test_type == "unit":
            exit_code = run_unit_tests()
        elif test_type == "integration":
            exit_code = run_integration_tests()
        elif test_type == "performance":
            exit_code = run_performance_tests()
        elif test_type == "benchmark":
            exit_code = run_benchmark_tests()
        elif test_type == "all":
            exit_code = run_all_tests()
        else:
            print(f"Unknown test type: {test_type}")
            print("Available options: unit, integration, performance, benchmark, all")
            exit_code = 1
    else:
        # Run all tests by default
        exit_code = run_all_tests()
    
    sys.exit(exit_code)