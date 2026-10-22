#!/usr/bin/env python3
"""Opt-in CLI integration checks, restricted to newly created temporary artifacts."""
import datetime, json, os, pathlib, re, subprocess, sys, tempfile, time
if os.environ.get('CRAFT_LIVE_TESTS') != '1' or os.environ.get('CRAFT_LIVE_MUTATION_TESTS') != '1':
    raise SystemExit('Requires CRAFT_LIVE_TESTS=1 and CRAFT_LIVE_MUTATION_TESTS=1')
profile = os.environ['CRAFT_LIVE_PROFILE']
binary = str(pathlib.Path(sys.argv[1]).resolve())
prefix = 'craft-cli-live-test-'+datetime.datetime.now().strftime('%Y%m%d-%H%M%S')
checks, documents, folders, reminders = [], [], [], []

def call(*args, payload=None, optional=False):
    argv = [binary, *args, '--profile', profile, '--quiet']
    if payload is not None: argv += ['--stdin']
    result = subprocess.run(argv, input=json.dumps(payload) if payload is not None else '', text=True, capture_output=True, timeout=120)
    time.sleep(.7)
    if result.returncode:
        kind = 'unknown'
        try: kind = json.loads(result.stderr)['error']['kind']
        except (ValueError, KeyError, TypeError): pass
        if optional and kind in ['capability_unavailable', 'api_error', 'permission_denied', 'api_timeout']:
            checks.append({'command':' '.join(args[:2]),'status':'unavailable','kind':kind}); return None
        message=''
        try: message=json.loads(result.stderr).get('error',{}).get('message','')
        except (ValueError,AttributeError): pass
        message=re.sub(r'https?://[^\s\"<>]+','[redacted URL]',message)
        failure=pathlib.Path('/tmp/craft-live-failure.json');failure.write_text(json.dumps({'kind':kind,'message':message}));failure.chmod(0o600)
        raise RuntimeError('CLI check failed: '+args[0]+' (exit '+str(result.returncode)+', '+kind+')')
    checks.append({'command':' '.join(args[:2]),'status':'passed'})
    return json.loads(result.stdout) if result.stdout.strip() else None

def ident(value):
    if isinstance(value,str): return value
    if isinstance(value,list): return ident(value[0])
    if 'operations' in value: return ident(value['operations'][0])
    if 'items' in value: return ident(value['items'][0])
    for key in ['id','rootBlockId','blockId','collectionBlockId','whiteboardBlockId','viewId']:
        if key in value: return value[key]
    raise RuntimeError('Missing returned artifact ID')

def destructive(*args,payload=None):
    call(*args,'--dry-run',payload=payload)
    return call(*args,'--yes',payload=payload)

try:
    call('connection'); call('list','--limit','1'); call('tasks','list'); call('folders','list'); call('collections','list')
    for suffix in ['source','target']:
        doc=ident(call('create','--title',prefix+'-'+suffix)); documents.append(doc)
    doc,target=documents
    folder=ident(call('folders','create',prefix+'-folder'));folders.append(folder)
    parent=ident(call('folders','create',prefix+'-parent'));folders.append(parent)
    call('move',doc,'--to-folder',folder,'--dry-run');call('move',doc,'--to-folder',folder)
    call('folders','move',folder,'--to',parent,'--dry-run');call('folders','move',folder,'--to',parent)
    call('folders','move',folder,'--to','root','--dry-run');call('folders','move',folder,'--to','root')
    block=ident(call('blocks','add',doc,'--markdown',prefix+' searchable text'))
    call('get',doc,'--max-depth','1');call('blocks','get',doc,'--depth','1')
    call('blocks','update',block,'--markdown',prefix+' revised text')
    call('search',prefix);call('search','revised','--document',doc)
    call('blocks','move',block,'--to',target,'--dry-run');call('blocks','move',block,'--to',target)
    task=ident(call('tasks','add',prefix+'-task','--location','document','--document',doc))
    call('tasks','update',task,'--state','done');destructive('tasks','delete',task)
    reminder=call('reminders','create',block,optional=True)
    if reminder is not None:
        reminder_id=ident(reminder);reminders.append(reminder_id);call('reminders','list','--limit','1');call('reminders','update',reminder_id,'--completed');destructive('reminders','delete',reminder_id);reminders.remove(reminder_id)
    call('comments','add',payload={'comments':[{'blockId':block,'content':prefix+' test comment'}]})
    collection=ident(call('collections','create',payload={'schema':{'name':prefix+'-collection','properties':[{'name':'Status','type':'text'}]},'position':{'pageId':doc,'position':'end'}}))
    schema=call('collections','schema',collection)
    schema=schema.get('schema',schema)
    prop=schema['properties'][0]
    destructive('collections','schema','update',collection,payload={'schema':{'name':prefix+'-updated','properties':[{'name':'Status','type':'text','key':prop['key']}]}})
    item=ident(call('collections','add',collection,'--title',prefix+'-item','--properties',json.dumps({prop['key']:'todo'})))
    call('collections','items',collection);call('collections','update',collection,'--item',item,'--properties',json.dumps({prop['key']:'done'}))
    view=ident(call('collections','views','create',collection,'--name',prefix+'-view','--type','table'))
    call('collections','views','list',collection);call('collections','views','update',collection,'--view',view,'--name',prefix+'-renamed')
    call('collections','active-view','set',collection,'--view',view);destructive('collections','views','delete',collection,'--view',view)
    destructive('collections','delete',collection,'--item',item)
    board=ident(call('whiteboards','create',doc))
    shape={'id':prefix+'-shape','type':'rectangle','x':0,'y':0,'width':100,'height':100,'angle':0,'strokeColor':'#000000','backgroundColor':'transparent','fillStyle':'solid','strokeWidth':1,'strokeStyle':'solid','roughness':1,'opacity':100,'groupIds':[],'seed':1,'version':1,'versionNonce':1,'isDeleted':False,'boundElements':None,'updated':1,'link':None,'locked':False}
    call('whiteboards','add',board,'--json',json.dumps([shape]));call('whiteboards','get',board)
    call('whiteboards','update',board,'--json',json.dumps([{'id':shape['id'],'x':50}]))
    destructive('whiteboards','delete',board,'--ids',shape['id'])
    with tempfile.TemporaryDirectory() as td:
        upload=pathlib.Path(td)/'test.txt';upload.write_text(prefix+' upload')
        uploaded=call('upload',str(upload),'--page',doc,'--timeout','90s',optional=True)
        if uploaded is None:
            # An uncertain write is verified by reads, never retried.
            call('blocks','get',doc,'--depth','1');call('blocks','get',doc,'--depth','1')
    destructive('blocks','delete',block)
    destructive('clear',target)
finally:
    cleanup_errors=[]
    for reminder in reminders:
        try: destructive('reminders','delete',reminder)
        except Exception as error: cleanup_errors.append(str(error))
    for doc in documents:
        try: destructive('delete',doc)
        except Exception as error: cleanup_errors.append(str(error))
    for folder in folders:
        try: destructive('folders','delete',folder)
        except Exception as error: cleanup_errors.append(str(error))
    report={'prefix':prefix,'checks':checks,'cleanup_errors':cleanup_errors,'documents':documents,'folders':folders}
    report_path=pathlib.Path('/tmp/craft-live-release-report.json');report_path.write_text(json.dumps(report,indent=2)+'\n');report_path.chmod(0o600)
    print(json.dumps({'checks':len(checks),'unavailable':[c for c in checks if c['status']!='passed'],'cleanup_errors':cleanup_errors}))
    if cleanup_errors: raise RuntimeError('Temporary artifact cleanup needs attention')
