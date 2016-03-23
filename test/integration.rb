#!/usr/bin/env ruby

require 'securerandom'
require 'tempfile'
require 'test/unit'

TEMP_PREFIX = 'svfs-test.'

class TestIntegration < Test::Unit::TestCase

  def setup
    @mountpoint = ENV['TEST_MOUNTPOINT']
    @new_name = "#{@mountpoint}/#{SecureRandom.hex}"
  end

  def test_empty_file
    # Create
    file = Tempfile.new(TEMP_PREFIX, @mountpoint)

    # Rename
    file.close()
    File.rename(file.path, @new_name)

    # Remove
    File.delete(@new_name)
  end

  def test_standard_file
    # Create
    file = Tempfile.new(TEMP_PREFIX, @mountpoint)

    # Write
    ENV['TEST_NSEG_SIZE'].to_i.times do |iter|
      file.write(SecureRandom.random_bytes(2**20))
    end

    # Make sure size is correct
    assert_equal(file.size, (2**20)*ENV['TEST_NSEG_SIZE'].to_i, "Wrong file size")

    # Rename
    file.close()
    File.rename(file.path, @new_name)

    # Remove
    File.delete(@new_name)
  end

  def test_segment
    # Create
    file = Tempfile.new(TEMP_PREFIX, @mountpoint)

    # Write
    ENV['TEST_SEG_SIZE'].to_i.times do |iter|
      file.write(SecureRandom.random_bytes(2**20))
    end

    # Make sure size is correct
    assert_equal(file.size, (2**20)*ENV['TEST_SEG_SIZE'].to_i, "Wrong file size")

    # Close
    file.close()
    sleep 5

    # Rename
    File.rename(file.path, @new_name)

    # Remove
    file.unlink()
    File.delete(@new_name)
  end

  def test_directory
    # Create
    Dir.mkdir(@new_name)

    # Remove
    Dir.rmdir(@new_name)
  end

end
