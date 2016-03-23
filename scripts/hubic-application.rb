#!/usr/bin/env ruby


# *****************************************************************************
#  SVFS: The Swift Virtual File System
# *****************************************************************************
#  SVFS allows mounting Swift storage as a file system, using fuse. Check the
#  project homepage for details and limitations to this approach.
# *****************************************************************************
#  @vendor : OVH
# *****************************************************************************


require 'base64'
require 'cgi'
require 'io/console'
require 'json'
require 'net/http'
require 'securerandom'


# Constants
HUBIC_URI = URI("https://api.hubic.com")
REDIRECT_URI = CGI.escape('http://localhost/')

# Make sure user has an application ready to use
print "Did you registered an application under your hubic account ? (y/N) "

unless STDIN.gets.chomp == 'y'
  puts "Please visit https://hubic.com/home/browser/developers/ and add an application."
  puts "Application names must be unique accross hubic, you could for instance use svfs-#{SecureRandom.uuid} for this."
  abort
end


client_id     = [(print " ~> Application client_id: "), STDIN.gets.rstrip][1]
client_secret = [(print " ~> Application client_secret: "), STDIN.noecho(&:gets).rstrip][1]

# Common
authorization_basic = Base64.strict_encode64("#{client_id}:#{client_secret}")
http = Net::HTTP.new(HUBIC_URI.host, HUBIC_URI.port)
http.use_ssl = (HUBIC_URI.port == 443)


#---------------------------------------
# STEP 1: TAILOR APPLICATION SCOPE
#---------------------------------------
print "\n1) Setting scope ... "

uri = "/oauth/auth/"\
  "?client_id=#{client_id}"\
  "&redirect_uri=#{REDIRECT_URI}"\
  "&scope=credentials.r"\
  "&response_type=code"\
  "&state=#{SecureRandom.base64(27)}"

response = http.get(uri)

if response.code != "200"
  puts "FAILED"
  abort "Can't access application authorization service (wrong client id ?)"
end

/name="oauth" value="(?<oauth_code>\d+)"/ =~ response.body

if oauth_code != nil && !oauth_code.empty?
  puts "OK"
else
  puts "FAILED"
  abort "Can't find oauth code in response from server !"
end


#---------------------------------------
# STEP 2: GRANT APPLICATION ACCESS
#---------------------------------------
uri   = "/oauth/auth/"

data  = {
  'action'        => 'accepted',
  'credentials'   => 'r',
  'oauth'         => oauth_code,
}

data['login']     = [(print " ~> Email: "), STDIN.gets.rstrip][1]
data['user_pwd']  = [(print " ~> Password: "), STDIN.noecho(&:gets).rstrip][1]

print "\n2) Granting access ... "

response = http.post(uri, URI.encode_www_form(data))

if response.code != "302"
  abort "Failed to authenticate this application (wrong email or password ?)"
end

/code=(?<author_code>\w+)/ =~ response["Location"]

if author_code != nil && !author_code.empty?
  puts "OK"
else
  puts "FAILED"
  abort "Can't find author code in response from server !"
end


#---------------------------------------
# STEP 3: GET REFRESH TOKEN
#---------------------------------------
print "3) Getting refresh token ... "

uri = "/oauth/token"

data = {
  'code'          => author_code,
  'redirect_uri'  => REDIRECT_URI,
  'grant_type'    => 'authorization_code',
}

response = http.post(uri, URI.encode_www_form(data), {"Authorization" => "Basic #{authorization_basic}"})

if response.code != "200"
  puts "FAILED"
  abort "Could not fetch access token from server ! (wrong client secret ?)"
end

tokens = JSON.parse(response.body)
refresh_token = tokens['refresh_token']

if refresh_token != nil && !refresh_token.empty?
  puts "OK"
else
  puts "FAILED"
  abort "Invalid json from server !"
end

puts "\n == Your mount options =="
puts " ~> hubic_auth=#{authorization_basic}"
puts " ~> hubic_token=#{refresh_token}"
