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

  # This test :
  # - creates an empty file using a unique name
  # - closes this file
  # - renames this file to somehting else (random)
  # - removes the new named file
  def test_empty_file
    file = Tempfile.new(TEMP_PREFIX, @mountpoint)
    file.close()
    File.rename(file.path, @new_name)
    File.delete(@new_name)
  end

  # This test :
  # - creates an empty file using a unique name
  # - writes TEST_NSEG_SIZE * 1MB of data in it
  # - closes this file
  # - renames this file to somehting else (random)
  # - removes the new named file
  def test_standard_file
    file = Tempfile.new(TEMP_PREFIX, @mountpoint)

    ENV['TEST_NSEG_SIZE'].to_i.times do |iter|
      file.write(SecureRandom.random_bytes(2**20))
    end

    assert_equal(file.size, (2**20)*ENV['TEST_NSEG_SIZE'].to_i, "Wrong file size")

    file.close()
    File.rename(file.path, @new_name)
    File.delete(@new_name)
  end

  # This test :
  # - creates an empty file using a unique name
  # - writes TEST_SEG_SIZE * 1MB of data in it
  # - closes this file
  # - renames this file to somehting else (random)
  # - removes the new named file
  def test_segment
    file = Tempfile.new(TEMP_PREFIX, @mountpoint)

    ENV['TEST_SEG_SIZE'].to_i.times do |iter|
      file.write(SecureRandom.random_bytes(2**20))
    end

    assert_equal(file.size, (2**20)*ENV['TEST_SEG_SIZE'].to_i, "Wrong file size")

    file.close()
    sleep 5
    File.rename(file.path, @new_name)
    file.unlink()
    File.delete(@new_name)
  end

  # This test :
  # - creates an empty directory with a random name
  # - removes it
  def test_directory
    Dir.mkdir(@new_name)
    Dir.rmdir(@new_name)
  end

end
