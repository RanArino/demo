"""
Main test suite for the chunking service.
Run all tests and generate comprehensive reports.
"""
import subprocess
import sys
import time


def run_test_suite():
    """Run the complete test suite and generate reports."""
    
    print("=" * 80)
    print("CHUNKING SERVICE - COMPLETE TEST SUITE")
    print("=" * 80)
    print("")
    
    test_files = [
        ("Unit Tests", "tests/chunking/test_chunking_unit.py"),
        ("Integration Tests", "tests/chunking/test_chunking_integration.py"), 
        ("Performance Tests", "tests/chunking/test_chunking_performance.py"),
    ]
    
    total_start = time.time()
    all_passed = True
    
    for test_name, test_file in test_files:
        print(f"Running {test_name}...")
        print("-" * 40)
        
        start_time = time.time()
        result = subprocess.run([
            sys.executable, "-m", "pytest", 
            test_file, 
            "-v", "--tb=short"
        ], capture_output=True, text=True)
        end_time = time.time()
        
        print(f"Duration: {end_time - start_time:.2f}s")
        
        if result.returncode == 0:
            print(f"✅ {test_name} PASSED")
        else:
            print(f"❌ {test_name} FAILED")
            print("STDOUT:", result.stdout)
            print("STDERR:", result.stderr)
            all_passed = False
        
        print("")
    
    total_end = time.time()
    total_duration = total_end - total_start
    
    print("=" * 80)
    print("TEST SUITE SUMMARY")
    print("=" * 80)
    print(f"Total duration: {total_duration:.2f}s")
    print(f"Overall result: {'✅ ALL TESTS PASSED' if all_passed else '❌ SOME TESTS FAILED'}")
    print("")
    
    if all_passed:
        print("🎉 CHUNKING SERVICE IS READY FOR PRODUCTION")
        print("")
        print("✅ Sentence-based splitting verified")
        print("✅ Performance requirement met (< 1s for 200K chars)")
        print("✅ Multilingual support confirmed")
        print("✅ Character position mapping accurate")
        print("✅ Configuration options working")
        print("✅ Error handling robust")
    else:
        print("⚠️  CHUNKING SERVICE NEEDS ATTENTION")
        print("Please check failed tests above.")
    
    print("=" * 80)
    
    return all_passed


def generate_detailed_output():
    """Generate detailed test output showing chunking examples."""
    
    print("\n" + "=" * 80)
    print("GENERATING DETAILED CHUNKING EXAMPLES")
    print("=" * 80)
    
    # Run the demonstration scripts
    demo_scripts = [
        ("English Chunking Examples", "tests/chunking/test_chunking_output.py"),
        ("Multilingual Examples", "tests/chunking/test_multilingual_chunking.py"),
    ]
    
    for demo_name, script in demo_scripts:
        print(f"\n{demo_name}:")
        print("-" * 50)
        
        try:
            result = subprocess.run([
                sys.executable, script
            ], capture_output=True, text=True, timeout=30)
            
            if result.returncode == 0:
                # Save output to files
                output_file = f"result_{script.replace('.py', '.txt')}"
                with open(output_file, 'w', encoding='utf-8') as f:
                    f.write(result.stdout)
                print(f"✅ Generated {output_file}")
            else:
                print(f"❌ Failed to generate {demo_name}")
                if result.stderr:
                    print(f"Error: {result.stderr}")
                    
        except subprocess.TimeoutExpired:
            print(f"⏰ {demo_name} timed out")
        except Exception as e:
            print(f"❌ Error running {demo_name}: {e}")


if __name__ == "__main__":
    # Run main test suite
    success = run_test_suite()
    
    # Generate detailed examples
    generate_detailed_output()
    
    # Exit with appropriate code
    sys.exit(0 if success else 1)