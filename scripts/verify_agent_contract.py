#!/usr/bin/env python3
"""Offline subprocess proof. Every network operation targets a loopback fixture."""
import http.server, json, os, pathlib, subprocess, sys, tempfile, threading, urllib.parse
binary = str(pathlib.Path(sys.argv[1]).resolve())
requests = []
status = 200
class Handler(http.server.BaseHTTPRequestHandler):
    def log_message(self, *args): pass
    def do_GET(self): self.reply()
    def do_POST(self): self.reply()
    def do_PUT(self): self.reply()
    def do_DELETE(self): self.reply()
    def reply(self):
        body=self.rfile.read(int(self.headers.get('Content-Length',0)))
        requests.append((self.command,self.path,body))
        p=urllib.parse.urlparse(self.path).path
        value={'items':[{'id':'fixture-a','title':'Alpha'},{'id':'fixture-b','title':'Beta'}]}
        if p=='/blocks': value={'id':'fixture-a','type':'page','title':{'value':'Alpha'},'content':[{'id':'block-a','type':'text','markdown':'Hello'}]}
        if p=='/folders':value={'items':[{'id':'folder-a','name':'Folder','folders':[{'id':'child','name':'Child'}]}]}
        if p=='/tasks':value={'items':[{'id':'task-a','markdown':'Task','taskInfo':{'state':'todo'},'location':{'type':'document','documentId':'fixture-a'}}]}
        if p=='/documents/search':value={'items':[{'documentId':'fixture-a','markdown':'Alpha'}]}
        if p=='/mcp':
            rpc=json.loads(body)
            if rpc['method']=='notifications/initialized': self.send_response(202);self.end_headers();return
            value={'jsonrpc':'2.0','id':rpc['id'],'result':{'isError':True,'content':[{'type':'text','text':'Fixture failure'}]}}
        if p=='/mcp' and rpc['method']=='initialize':value['result']={'protocolVersion':'2025-06-18','capabilities':{},'serverInfo':{'name':'fixture','version':'1'}}
        if self.command=='DELETE':value={'items':[{'id':'fixture-a','status':'deleted'},{'id':'fixture-b','error':'fixture refusal'}]}
        if status!=200:value={'message':'Fixture failure'}
        self.send_response(status);self.send_header('Content-Type','application/json');self.send_header('Retry-After','12');self.send_header('X-RateLimit-Scope','space');self.end_headers();self.wfile.write(json.dumps(value).encode())
server=http.server.ThreadingHTTPServer(('127.0.0.1',0),Handler)
threading.Thread(target=server.serve_forever,daemon=True).start()
url='http://127.0.0.1:'+str(server.server_port)
results=[]
def run(args, expected=0, remote=True, cwd=None, stdin='', env=None):
    full=args+(['--api-url',url,'--api-key','fixture-key'] if remote else [])
    start=len(requests)
    p=subprocess.run([binary]+full,input=stdin,text=True,capture_output=True,timeout=8,cwd=cwd,env=env)
    assert p.returncode==expected,(args,p.returncode,p.stderr)
    results.append({'args':args,'exit':p.returncode,'requests':len(requests)-start})
    return p,requests[start:]
def data(args,**kw):p,r=run(args,**kw);return json.loads(p.stdout),r
manifest,_=data(['schema'],remote=False)
assert manifest['clispec']=='0.2'
for command in manifest['commands']:
    run(command['name'].split()+['--help'],remote=False)
    assert command['effects'] in ['read_only','idempotent','non_idempotent']
for args in [['tasks','add','--request-schema'],['collections','views','create','--response-schema'],['reminders','update','--request-schema']]:
    result,r=data(args,remote=False);assert not r and '$schema' in result
for args in [['list'],['folders','list'],['tasks','list']]:result,_=data(args);assert isinstance(result['items'],list)
result,_=data(['list','--count']);assert result['count']==2
p,_=run(['list','--id-only']);assert p.stdout.splitlines()==['fixture-a','fixture-b']
result,_=data(['list','--transform','items.0.id']);assert result=='fixture-a'
p,_=run(['list','--format','jsonl']);assert len(p.stdout.splitlines())==2;[json.loads(line) for line in p.stdout.splitlines()]
p,_=run(['list','--format','yaml']);assert 'items:' in p.stdout
p,r=run(['list','--format','bad'],expected=1);assert not r
p,r=run(['list','--deliver','bad:target'],expected=1);assert not r
p,r=run(['list','--data-source','local'],expected=1);assert not r
p,r=run(['delete','fixture-a'],expected=1);assert not r and '--yes' in p.stderr
p,r=run(['delete','fixture-a','--yes'],expected=2);assert json.loads(p.stdout)['items'][1]['error']=='fixture refusal'
for args in [['tasks','add','Fixture'],['folders','create','Fixture'],['delete','fixture-a'],['blocks','move','block-a','--to','fixture-a'],['reminders','create','block-a']]:
    # Block move flag is checked separately by its API fixture.
    if args[0]=='blocks':continue
    result,r=data(args+['--dry-run']);assert result['dry_run'] and result['validated']=='local';assert all(x[0]=='GET' for x in r)
p,r=run(['mcp','call','craft_write','--command','documents create --title Fixture','--mcp-url',url+'/mcp','--dry-run'],remote=False);assert not r
p,r=run(['mcp','call','craft_write','--command','documents create --title Fixture','--mcp-url',url+'/mcp'],remote=False,expected=1);assert not r
p,r=run(['mcp','call','craft_read','--command','documents list','--mcp-url',url+'/mcp'],remote=False,expected=2);assert json.loads(p.stderr)['code']=='MCP_TOOL_ERROR'
result,r=data(['search','Alpha','--folder','one,two']);q=urllib.parse.parse_qs(urllib.parse.urlparse(r[0][1]).query);assert q['folderIds']==['one','two'] and 'folderIDs' not in q
result,r=data(['get','fixture-a','--max-depth','1']);assert 'maxDepth=1' in r[0][1] and result['title']=='Alpha'
for status in [401,404,429,500]:
    p,_=run(['list'],expected=2);err=json.loads(p.stderr);assert err['error']['kind']==err['code'].lower()
    if status==429:assert err['details']['Retry-After']=='12' and err['details']['X-Ratelimit-Scope']=='space'
status=200
with tempfile.TemporaryDirectory() as td:
    target=pathlib.Path(td)/'out.json'
    p,_=run(['list','--deliver','file:out.json'],cwd=td);assert not p.stdout and json.loads(target.read_text())['total']==2
    p,r=run(['list','--deliver','file:out.json'],cwd=td,expected=1);assert not r
    result,_=data(['skill-path'],remote=False,cwd=td);assert result.get('content') and 'untrusted' in result['content']
with tempfile.TemporaryDirectory() as td:
    env=os.environ.copy();env['CRAFT_CONFIG_DIR']=td;env['CRAFT_FIXTURE_SECRET']='fixture-private-value'
    config=pathlib.Path(td)/'config.json';config.write_text('{invalid')
    result,_=data(['schema'],remote=False,env=env);assert result['clispec']=='0.2'
    config.unlink()
    p,r=run(['profiles','add-rest','fixture','--api-url',url,'--api-key-env','CRAFT_FIXTURE_SECRET','--dry-run'],remote=False,env=env);assert not config.exists()
    run(['profiles','add-rest','fixture','--api-url',url,'--api-key-env','CRAFT_FIXTURE_SECRET'],remote=False,env=env)
    assert config.stat().st_mode & 0o777 == 0o600
    p,_=run(['profiles','show','fixture'],remote=False,env=env);assert 'fixture-private-value' not in p.stdout
    p,_=run(['profiles','list'],remote=False,env=env);assert json.loads(p.stdout)['source']=='config_file'
    config.unlink();p,r=run(['list'],remote=False,env=env,expected=3);assert not p.stdout and not r
repo=pathlib.Path(__file__).resolve().parents[1]
assert (repo/'SKILL.md').read_text()==(repo/'internal/agentdoc/SKILL.md').read_text(), 'bundled skill drift'
assert json.loads((repo/'docs/command-reference.json').read_text())==manifest,'command reference drift'
p,r=run(['setup'],remote=False,expected=1);assert not p.stdout and not r
print(json.dumps({'status':'passed','checks':len(results),'production_mutations':0,'results':results},indent=2))
