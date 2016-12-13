require 'resolv'
require 'aws-sdk'
require 'net/http'
require 'json'
require 'json-compare'
require 'yajl'
require 'open-uri'

# Write config
def write_config(config_file, url)
  open(config_file, 'wb') do |file|
    open(url) do |uri|
      file.write(uri.read)
    end
  end
end

# Compare json
def compare(j1, j2)
  # j1 = Yajl::Parser.parse(j1)
  # j2 = Yajl::Parser.parse(j2)
  JsonCompare.get_diff(j2, j1)
end

# Variables
metadata_endpoint = 'http://169.254.169.254/latest/dynamic/instance-identity/document'
instance_data = JSON.parse(Net::HTTP.get(URI.parse(metadata_endpoint)))
config_values = {}

# Connection to Amazon
ec2 = Aws::EC2::Client.new(region: instance_data['region'])

# Get tags from instance
resp = ec2.describe_tags(filters: [{ name: 'resource-id', values: [instance_data['instanceId']] }]).to_h

# Put discovery url in config_values hash
resp[:tags].each { |h| config_values[:url] = h[:value] if h[:key] == 'Discovery_url' }

# Quit if we cant find the value
abort 'No value found' if config_values[:url].nil?

# Find the data needed for whatever
dns = Resolv::DNS.new
dns.getresources(config_values[:url], Resolv::DNS::Resource::IN::TXT).collect do |r|
  r = r.data.split('=')
  config_values[r[0]] = r[1]
end

# Variables again
url = config_values['s3_url']
file = '/etc/logstash/conf.d/logstash.conf'

if File.file?(file)
  new_file = open(url).read
  old_file = File.new(file, 'r')
  write_config(file, url) if compare(old_file, new_file) == false
else
  write_config(file, url)
end

# TODO: Discover usage
# TODO: Move config of logstash and not grok to s3 + variables??????
