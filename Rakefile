# coding: utf-8
require 'rake'

# Tool Versions
NODE_VER = '12.13.1'
SWAGGER_CODEGEN_VER = '2.4.12'
GOSWAGGER_VER = 'v0.21.0'
GOLANGCILINT_VER = '1.21.0'
GO_VER = '1.13.5'
PROTOC_VER = '3.11.2'
PROTOC_GEN_GO_VER = 'v1.3.3'

# Check host OS
UNAME=`uname -s`

case UNAME.rstrip
  when "Darwin"
    OS="macos"
    GOSWAGGER_BIN="swagger_darwin_amd64"
    GO_SUFFIX="darwin-amd64"
    PROTOC_ZIP_SUFFIX="osx-x86_64"
    NODE_SUFFIX="darwin-x64"
    GOLANGCILINT_SUFFIX="darwin-amd64"
    puts "WARNING: MacOS is not officially supported, the provisions for building on MacOS are made"
    puts "WARNING: for the developers' convenience only."
  when "Linux"
    OS="linux"
    GOSWAGGER_BIN="swagger_linux_amd64"
    GO_SUFFIX="linux-amd64"
    PROTOC_ZIP_SUFFIX="linux-x86_64"
    NODE_SUFFIX="linux-x64"
    GOLANGCILINT_SUFFIX="linux-amd64"
  when "FreeBSD"
    OS="FreeBSD"
    # TODO: there are no swagger built packages for FreeBSD
    GOSWAGGER_BIN=""
    puts "WARNING: There are no FreeBSD packages for GOSWAGGER_BIN"
    GO_SUFFIX="freebsd-amd64"
    # TODO: there are no protoc built packages for FreeBSD (at least as of 3.10.0)
    PROTOC_ZIP_SUFFIX=""
    puts "WARNING: There are no protoc packages built for FreeBSD"
    NODE_SUFFIX="node-v10.16.3.tar.xz"
    GOLANGCILINT_SUFFIX="freebsd-amd64"
  else
    puts "ERROR: Unknown/unsupported OS: %s" % UNAME
    fail
  end

# Tool URLs
GOSWAGGER_URL = "https://github.com/go-swagger/go-swagger/releases/download/#{GOSWAGGER_VER}/#{GOSWAGGER_BIN}"
GOLANGCILINT_URL = "https://github.com/golangci/golangci-lint/releases/download/v#{GOLANGCILINT_VER}/golangci-lint-#{GOLANGCILINT_VER}-#{GOLANGCILINT_SUFFIX}.tar.gz"
GO_URL = "https://dl.google.com/go/go#{GO_VER}.#{GO_SUFFIX}.tar.gz"
PROTOC_URL = "https://github.com/protocolbuffers/protobuf/releases/download/v#{PROTOC_VER}/protoc-#{PROTOC_VER}-#{PROTOC_ZIP_SUFFIX}.zip"
PROTOC_GEN_GO_URL = 'github.com/golang/protobuf/protoc-gen-go'
SWAGGER_CODEGEN_URL = "https://oss.sonatype.org/content/repositories/releases/io/swagger/swagger-codegen-cli/#{SWAGGER_CODEGEN_VER}/swagger-codegen-cli-#{SWAGGER_CODEGEN_VER}.jar"
NODE_URL = "https://nodejs.org/dist/v#{NODE_VER}/node-v#{NODE_VER}-#{NODE_SUFFIX}.tar.xz"
MOCKERY_URL = 'github.com/vektra/mockery/.../'
MOCKGEN_URL = 'github.com/golang/mock/mockgen'
RICHGO_URL = 'github.com/kyoh86/richgo'

# Tools and Other Paths
TOOLS_DIR = File.expand_path('tools')
NPX = "#{TOOLS_DIR}/node-v#{NODE_VER}-#{NODE_SUFFIX}/bin/npx"
SWAGGER_CODEGEN = "#{TOOLS_DIR}/swagger-codegen-cli-#{SWAGGER_CODEGEN_VER}.jar"
GOSWAGGER_DIR = "#{TOOLS_DIR}/#{GOSWAGGER_VER}"
GOSWAGGER = "#{GOSWAGGER_DIR}/#{GOSWAGGER_BIN}"
NG = File.expand_path('webui/node_modules/.bin/ng')
GOHOME_DIR = File.expand_path('~/go')
GOBIN = "#{GOHOME_DIR}/bin"
GO_DIR = "#{TOOLS_DIR}/#{GO_VER}"
GO = "#{GO_DIR}/go/bin/go"
GOLANGCILINT = "#{TOOLS_DIR}/golangci-lint-#{GOLANGCILINT_VER}-#{GOLANGCILINT_SUFFIX}/golangci-lint"
PROTOC_DIR = "#{TOOLS_DIR}/#{PROTOC_VER}"
PROTOC = "#{PROTOC_DIR}/bin/protoc"
PROTOC_GEN_GO = "#{GOBIN}/protoc-gen-go-#{PROTOC_GEN_GO_VER}"
MOCKERY = "#{GOBIN}/mockery"
MOCKGEN = "#{GOBIN}/mockgen"
RICHGO = "#{GOBIN}/richgo"

# Patch PATH env
ENV['PATH'] = "#{TOOLS_DIR}/node-v#{NODE_VER}-#{NODE_SUFFIX}/bin:#{ENV['PATH']}"
ENV['PATH'] = "#{GO_DIR}/go/bin:#{ENV['PATH']}"
ENV['PATH'] = "#{GOBIN}:#{ENV['PATH']}"

# build date
build_date = Time.now.strftime("%Y-%m-%d %H:%M")
puts "Stork build date: #{build_date}"
go_build_date_opt = "-ldflags=\"-X 'isc.org/stork.BuildDate=#{build_date}'\""

# Documentation
SPHINXOPTS = "-v -E -a -W -j 2"

# Files
SWAGGER_FILE = File.expand_path('api/swagger.yaml')
SWAGGER_API_FILES = [
  'api/swagger.in.yaml',
  'api/services-defs.yaml', 'api/services-paths.yaml',
  'api/users-defs.yaml', 'api/users-paths.yaml',
  'api/dhcp-defs.yaml', 'api/dhcp-paths.yaml'
]
AGENT_PROTO_FILE = File.expand_path('backend/api/agent.proto')
AGENT_PB_GO_FILE = File.expand_path('backend/api/agent.pb.go')

SERVER_GEN_FILES = Rake::FileList[
  File.expand_path('backend/server/gen/restapi/configure_stork.go'),
]

# Directories
directory GOHOME_DIR
directory TOOLS_DIR


# Server Rules
file GO => [TOOLS_DIR, GOHOME_DIR] do
  sh "mkdir -p #{GO_DIR}"
  sh "wget #{GO_URL} -O #{GO_DIR}/go.tar.gz"
  Dir.chdir(GO_DIR) do
    sh 'tar -zxf go.tar.gz'
  end
end

YAMLINC = File.expand_path('webui/node_modules/.bin/yamlinc')

file YAMLINC do
  Rake::Task[NG].invoke()
end

file SWAGGER_FILE => [YAMLINC, *SWAGGER_API_FILES] do
  Dir.chdir('api') do
    sh "#{YAMLINC} -o swagger.yaml swagger.in.yaml"
  end
end

file SERVER_GEN_FILES => SWAGGER_FILE do
  Dir.chdir('backend') do
    sh "#{GOSWAGGER} generate server -s server/gen/restapi -m server/gen/models --name Stork --exclude-main --spec #{SWAGGER_FILE} --template stratoscale --regenerate-configureapi"
  end
end

desc 'Generate server part of REST API using goswagger based on swagger.yml'
task :gen_server => [GO, GOSWAGGER, SERVER_GEN_FILES]

file GOSWAGGER => TOOLS_DIR do
  sh "mkdir -p #{GOSWAGGER_DIR}"
  sh "wget #{GOSWAGGER_URL} -O #{GOSWAGGER}"
  sh "chmod a+x #{GOSWAGGER}"
end

desc 'Compile server part'
task :build_server => [GO, :gen_server, :gen_agent] do
  sh 'rm -f backend/server/agentcomm/api_mock.go'
  sh "cd backend/cmd/stork-server/ && #{GO} build #{go_build_date_opt}"
end

file PROTOC do
  sh "mkdir -p #{PROTOC_DIR}"
  sh "wget #{PROTOC_URL} -O #{PROTOC_DIR}/protoc.zip"
  Dir.chdir(PROTOC_DIR) do
    sh 'unzip protoc.zip'
  end
end

file PROTOC_GEN_GO do
  sh "#{GO} get -d -u #{PROTOC_GEN_GO_URL}"
  sh "git -C \"$(#{GO} env GOPATH)\"/src/github.com/golang/protobuf checkout #{PROTOC_GEN_GO_VER}"
  sh "#{GO} install github.com/golang/protobuf/protoc-gen-go"
  sh "cp #{GOBIN}/protoc-gen-go #{PROTOC_GEN_GO}"
end

file MOCKERY do
  sh "#{GO} get -u #{MOCKERY_URL}"
end

file MOCKGEN do
  sh "#{GO} get -u #{MOCKGEN_URL}"
end

file RICHGO do
  sh "#{GO} get -u #{RICHGO_URL}"
end

file AGENT_PB_GO_FILE => [GO, PROTOC, PROTOC_GEN_GO, AGENT_PROTO_FILE] do
  Dir.chdir('backend') do
    sh "#{PROTOC} -I api api/agent.proto --go_out=plugins=grpc:api"
  end
end

# prepare args for dlv debugger
headless = ''
if ENV['headless'] == 'true'
  headless = '--headless -l 0.0.0.0:45678'
end

desc 'Connect gdlv GUI Go debugger to waiting dlv debugger'
task :connect_dbg do
  sh 'gdlv connect 127.0.0.1:45678'
end

desc 'Generate API sources from agent.proto'
task :gen_agent => [AGENT_PB_GO_FILE]

desc 'Compile agent part'
file :build_agent => [GO, AGENT_PB_GO_FILE] do
  sh "cd backend/cmd/stork-agent/ && #{GO} build #{go_build_date_opt}"
end

desc 'Run agent'
task :run_agent => [:build_agent, GO] do
  if ENV['debug'] == 'true'
    sh "cd backend/cmd/stork-agent/ && dlv #{headless} debug"
  else
    sh "backend/cmd/stork-agent/stork-agent --port 8888"
  end
end

desc 'Run server'
task :run_server => [:build_server, GO] do |t, args|
  if ENV['debug'] == 'true'
    sh "cd backend/cmd/stork-server/ && dlv #{headless} debug"
  else
    cmd = 'backend/cmd/stork-server/stork-server'
    if ENV['dbtrace'] == 'true'
      cmd = "#{cmd} --db-trace-queries"
    end
    sh cmd
  end
end

desc 'Run server with local postgres docker container'
task :run_server_db do |t, args|
  ENV['STORK_DATABASE_NAME'] = "storkapp"
  ENV['STORK_DATABASE_USER_NAME'] = "storkapp"
  ENV['STORK_DATABASE_PASSWORD'] = "storkapp"
  ENV['STORK_DATABASE_HOST'] = "localhost"
  ENV['STORK_DATABASE_PORT'] = "5678"
  at_exit {
    sh "docker rm -f -v stork-app-pgsql"
  }
  sh 'docker run --name stork-app-pgsql -d -p 5678:5432 -e POSTGRES_DB=storkapp -e POSTGRES_USER=storkapp -e POSTGRES_PASSWORD=storkapp postgres:11 && sleep 5'
  Rake::Task["run_server"].invoke()
end


desc 'Compile database migrations tool'
task :build_migrations =>  [GO] do
  sh "cd backend/cmd/stork-db-migrate/ && #{GO} build #{go_build_date_opt}"
end

desc 'Compile whole backend: server, migrations and agent'
task :build_backend => [:build_agent, :build_server, :build_migrations]

file GOLANGCILINT => TOOLS_DIR do
  Dir.chdir(TOOLS_DIR) do
    sh "wget #{GOLANGCILINT_URL} -O golangci-lint.tar.gz"
    sh "tar -zxf golangci-lint.tar.gz"
  end
end

desc 'Check backend source code
arguments: fix=true - fixes some of the found issues'
task :lint_go => [GO, GOLANGCILINT, MOCKERY, MOCKGEN, :gen_agent, :gen_server] do
  at_exit {
    sh 'rm -f backend/server/agentcomm/api_mock.go'
  }
  sh 'rm -f backend/server/agentcomm/api_mock.go'
  Dir.chdir('backend') do
    sh "#{GO} generate -v ./..."

    opts = ''
    if ENV['fix'] == 'true'
      opts += ' --fix'
    end
    sh "#{GOLANGCILINT} run #{opts}"
  end
end

desc 'Format backend source code'
task :fmt_go => [GO, :gen_agent, :gen_server] do
  Dir.chdir('backend') do
    sh "#{GO} fmt ./..."
  end
end

desc 'Run backend unit tests'
task :unittest_backend => [GO, RICHGO, MOCKERY, MOCKGEN, :build_server, :build_agent] do
  at_exit {
    sh 'rm -f backend/server/agentcomm/api_mock.go'
  }
  sh 'rm -f backend/server/agentcomm/api_mock.go'
  if ENV['scope']
    scope = ENV['scope']
  else
    scope = './...'
  end
  Dir.chdir('backend') do
    sh "#{GO} generate -v ./..."
    if ENV['debug'] == 'true'
      sh "dlv #{headless} test #{scope}"
    else
      gotool = RICHGO
      if ENV['richgo'] == 'false'
        gotool = GO
      end
      sh "#{gotool} test -race -v -count=1 -p 1 -coverprofile=coverage.out  #{scope}"  # count=1 disables caching results
    end

    # check coverage level
    out = `#{GO} tool cover -func=coverage.out`
    puts out, ''
    problem = false
    out.each_line do |line|
      if line.start_with? 'total:' or line.include? 'api_mock.go'
        next
      end
      items = line.gsub(/\s+/m, ' ').strip.split(" ")
      file = items[0]
      func = items[1]
      cov = items[2].strip()[0..-2].to_f
      ignore_list = ['DetectServices', 'RestartKea', 'Serve', 'BeforeQuery', 'AfterQuery',
                     'Identity', 'LogoutHandler', 'NewDatabaseSettings', 'ConnectionParams',
                     'Password', 'loggingMiddleware', 'GlobalMiddleware', 'Authorizer',
                     'CreateSession', 'DeleteSession', 'Listen', 'Shutdown', 'NewRestUser',
                     'CreateUser', 'UpdateUser', 'SetupLogging', 'UTCNow', 'detectApps',
                     'prepareTLS']
      if cov < 35 and not ignore_list.include? func
        puts "FAIL: %-80s %5s%% < 35%%" % ["#{file} #{func}", "#{cov}"]
        problem = true
      end
    end
    if problem
      fail("\nFAIL: Tests coverage is too low, add some tests\n\n")
    end
  end
end

desc 'Run backend unit tests with local postgres docker container'
task :unittest_backend_db do
  at_exit {
    sh "docker rm -f -v stork-ut-pgsql"
  }
  sh "docker run --name stork-ut-pgsql -d -p 5678:5432 -e POSTGRES_DB=storktest -e POSTGRES_USER=storktest -e POSTGRES_PASSWORD=storktest postgres:11"
  ENV['POSTGRES_ADDR'] = "localhost:5678"
  Rake::Task["unittest_backend"].invoke
end

desc 'Show backend coverage of unit tests in web browser'
task :show_cov do
  at_exit {
    sh 'rm -f backend/server/agentcomm/api_mock.go'
  }
  if not File.file?('backend/coverage.out')
    Rake::Task["unittest_backend_db"].invoke()
  end
  Dir.chdir('backend') do
    sh "#{GO} generate -v ./..."
    sh "#{GO} tool cover -html=coverage.out"
  end
end


# Web UI Rules
desc 'Generate client part of REST API using swagger_codegen based on swagger.yml'
task :gen_client => [SWAGGER_CODEGEN, SWAGGER_FILE] do
  Dir.chdir('webui') do
    sh "java -jar #{SWAGGER_CODEGEN} generate -l typescript-angular -i #{SWAGGER_FILE} -o src/app/backend --additional-properties snapshot=true,ngVersion=8.2.8"
  end
end

file SWAGGER_CODEGEN => TOOLS_DIR do
  sh "wget #{SWAGGER_CODEGEN_URL} -O #{SWAGGER_CODEGEN}"
end

file NPX => TOOLS_DIR do
  Dir.chdir(TOOLS_DIR) do
    sh "wget #{NODE_URL} -O #{TOOLS_DIR}/node.tar.xz"
    sh "tar -Jxf node.tar.xz"
  end
end

file NG => NPX do
  Dir.chdir('webui') do
    sh 'npm install'
  end
end

desc 'Build angular application'
task :build_ui => [NG, :gen_client] do
  Dir.chdir('webui') do
    sh 'npx ng build --prod'
  end
end

desc 'Serve angular app'
task :serve_ui => [NG, :gen_client] do
  Dir.chdir('webui') do
    sh 'npx ng serve --disable-host-check --proxy-config proxy.conf.json'
  end
end

desc 'Check frontend source code'
task :lint_ui => [NG, :gen_client] do
  Dir.chdir('webui') do
    sh 'npx ng lint'
    sh 'npx prettier --config .prettierrc --check \'**/*\''
  end
end

desc 'Make frontend source code prettier'
task :prettier_ui => [NG, :gen_client] do
  Dir.chdir('webui') do
    sh 'npx prettier --config .prettierrc --write \'**/*\''
  end
end

# internal task used in ci for running npm ci command with lint and tests together
task :ci_ui => [:gen_client] do
  Dir.chdir('webui') do
    sh 'npm ci'
  end

  Rake::Task["lint_ui"].invoke()

#   Dir.chdir('webui') do
#    sh 'CHROME_BIN=/usr/bin/chromium-browser npx ng test --progress false --watch false'
#    sh 'npx ng e2e --progress false --watch false'
#   end
end


# Docker Rules
desc 'Build containers with everything and statup all services using docker-compose
arguments: cache=false - forces rebuilding whole container'
task :docker_up => [:build_backend, :build_ui] do
  at_exit {
    sh "docker-compose down"
  }
  cache_opt = ''
  if ENV['cache'] == 'false'
    cache_opt = '--no-cache'
  end
  sh "docker-compose build #{cache_opt}"
  sh 'docker-compose up'
end

desc 'Shut down all containers'
task :docker_down do
  sh 'docker-compose down'
end

desc 'Build container with Stork Agent and Kea DHCPv4 server'
task :build_kea_container do
  sh 'docker build -f docker/docker-agent-kea.txt -t agent-kea .'
end

desc 'Run container with Stork Agent and Kea and mount current Agent binary'
task :run_kea_container do
  # host[8888]->agent[8080],  host[8787]->kea-ca[8000]
  sh 'docker run --rm -ti -p 8888:8080 -p 8787:8000 -h agent-kea -v `pwd`/backend/cmd/stork-agent:/agent agent-kea'
end

desc 'Build container with Stork Agent and Kea DHCPv6 server'
task :build_kea6_container do
  sh 'docker build -f docker/docker-agent-kea6.txt -t agent-kea6 .'
end

desc 'Run container with Stork Agent and Kea DHCPv6 server and mount current Agent binary'
task :run_kea6_container do
  # host[8888]->agent[8080]
  sh 'docker run --rm -ti -p 8886:8080 -h agent-kea6 -v `pwd`/backend/cmd/stork-agent:/agent agent-kea6'
end

desc 'Build two containers with Stork Agent and Kea HA pair
arguments: cache=false - forces rebuilding whole container'
task :build_kea_ha_containers => :build_agent do
  cache_opt = ''
  if ENV['cache'] == 'false'
    cache_opt = '--no-cache'
  end
  sh "docker-compose build #{cache_opt} agent-kea-ha1 agent-kea-ha2"
end

desc 'Run two containers with Stork Agent and Kea HA pair'
task :run_kea_ha_containers do
  at_exit {
    sh "docker-compose down"
  }
  sh 'docker-compose up agent-kea-ha1 agent-kea-ha2'
end

desc 'Build container with Stork Agent and BIND 9'
task :build_bind9_container do
  sh 'docker build -f docker/docker-agent-bind9.txt -t agent-bind9 .'
end

desc 'Run container with Stork Agent and BIND 9 and mount current Agent binary'
task :run_bind9_container do
  # host[9999]->agent[8080]
  sh 'docker run --rm -ti -p 9999:8080 -h agent-bind9 -v `pwd`/backend/cmd/stork-agent:/agent agent-bind9'
end

desc 'Starts generating DHCP traffic (starts the traffic-dhcp docker container if it isn\'t running)'
task :start_traffic_dhcp do
  dhcp_traffic_id = `docker ps | grep traffic-dhcp | awk '{ print $1 }'`
  if dhcp_traffic_id != ""
    puts "traffic-dhcp container already running: #{dhcp_traffic_id}"
  else
    sh 'docker-compose up traffic-dhcp'
  end
end

desc 'Stops generating DHCP traffic (stops the traffic-dhcp docker container)'
task :stop_traffic_dhcp do
  dhcp_traffic_id = `docker ps | grep traffic-dhcp | awk '{ print $1 }'`
  if dhcp_traffic_id != ""
    sh "echo Stopping Docker container: #{dhcp_traffic_id}"
    sh "docker stop #{dhcp_traffic_id}"
  else
    puts "traffic-dhcp container not found"
  end
end


# Documentation
desc 'Builds Stork documentation, using Sphinx'
task :doc do
  sh "sphinx-build -M singlehtml doc/ doc/ #{SPHINXOPTS}"
end


# Release Rules
task :tarball do
  version = 'unknown'
  version_file = 'backend/version.go'
  text = File.open(version_file).read
  text.each_line do |line|
    if line.start_with? 'const Version'
      parts = line.split('"')
      version = parts[1]
    end
  end
  sh "git archive --prefix=stork-#{version}/ -o stork-#{version}.tar.gz HEAD"
end


# Other Rules
desc 'Remove tools and other build or generated files'
task :clean do
  sh "rm -rf #{AGENT_PB_GO_FILE}"
  sh 'rm -rf backend/server/gen/*'
  sh 'rm -rf webui/src/app/backend/'
  sh 'rm -f backend/cmd/stork-agent/stork-agent'
  sh 'rm -f backend/cmd/stork-server/stork-server'
  sh 'rm -f backend/cmd/stork-db-migrate/stork-db-migrate'
end

desc 'Download all dependencies'
task :prepare_env => [GO, GOSWAGGER, GOLANGCILINT, SWAGGER_CODEGEN, NPX] do
  sh "#{GO} get -u github.com/go-delve/delve/cmd/dlv"
  sh "#{GO} get -u github.com/aarzilli/gdlv"
end

desc 'Generate ctags for Emacs'
task :ctags do
  sh 'etags.ctags -f TAGS -R --exclude=webui/node_modules --exclude=tools .'
end
