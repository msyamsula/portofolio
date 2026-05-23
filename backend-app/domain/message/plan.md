I want to create create an isolated backend under backend-app/message folder

it has api to send message
I want to build event as event log concept

one event that should be atomic is within
- idempotencykey, conversationid, version -> this should be deterministically hash as (eventid). now eventid must unique across the record
- version is used for concurrency control everyone trying to append the log must query first to idempotencykey,conversationid, version pair, if not exist insert and use version = 0, get the version and when appending new event version must be add +1. so multiple process writing on this will be rejected, only one win


the message system will use outbox pattern so when appending to the log message, kafka will get the message and consume it, another consumer down the line must process the event if any

idempotencykey, conversationid, version must propagate within my system, because idempotencykey, conversationid, version build (eventid)

my distributed system will accept duplicate works and computation but the append log should be atomic by eventid and version guard the double book/insert

state should be constructed from log event, but it is an after thought, my ssource of truth is event log

this is the flow of my system


event log
- event id, idem key, conversationid, senderid, receiverid, version, retry, reconcile attempt, event name
eventid = hash(idemkey, senderid, receiverid, version): unique within the log

conversation
id:
senderid:
receiverid:
messageid:

rename idemkey = messageid form now


state machine
a: user click send
b: message is processed
c: message failed
d: message done

a - b: SENT
api commit to event log using eventid(hash), messageid, conversationid, version, this insert atomic (on conflict do nothing), if it is inserted then return early, start version from 0, retry, reconcile attempt = 0

if no conflict, commit it, on commit failed return faild in api

if success commit, outbox pattern will trigger SENT EVENT
message event: conversationid, messageid, version, message, senderid, receiverid

notes: make outbox pattern as appending in event log table for every event added
notes: save also event name to log for every work from now on
I will not always state both notes explicitly from now on


b - d: SUCCESS
get event log using conversationid, messageid
construct state
get conversationid, messageid, version, message, senderid, receiverid, retry
open trx:
- insert to event log with version++, eventid(new hash with new version), event success
- add to conversation table
- commit
if commit success outbox will work
if not let reconcile work for this


reconciler: a cron job work for every 1 minutes
- query all message of that has retry < code config> or failed < code config times (using event log construction): case 1
in case 1:
construct the state
if state = failed, event = RECONCILE FAILED MESSAGE
if state = processed, event = RECONCILE STUCKED MESSAGE

try this trx:
- append event log conversationid, version++, eventid, messageid, name above as event name
- add to conversation table
- commit
if commit success outbox will work
if not leave it cron will get this again

b - b: RECONCILE STUCKED MESSAGE
construct the state
if retry = max retry (code config):
open trx:
- append FAILED to event log, conversationid, mesageid, version++, attempt++
- commit
if commit success go to terminal failed state, will be swept by cron
if failed, it will be retried again

if not 
construct state
get conversationid, messageid, version, message, senderid, receiverid, retry
open trx:
- insert to event log with version++, eventid(new hash with new version)
- add to conversation table
- commit
if commit done
if not let reconcile work for this

b-b: RECONCILE FAILED MESSAGE
construct the state
if attempt = max attempt (code config):
open trx:
- append FAILED to event log, conversationid, mesageid, version++, attempt++
- commit
if commit success go to terminal failed state, will be swept by cron
if failed, it will be retried again

if not
construct state
get conversationid, messageid, version, message, senderid, receiverid, retry
open trx:
- insert to event log with version++, eventid(new hash with new version), success
- add to conversation table
- commit
if commit success done
if not let reconcile work for this

notice that reconcile attempt in event log is not needed anymore, retry will be share in RECONCILE FAILED MESSAGE, RECONCILE STUCKED MESSAGE


i think retry and attempt can be count from the event so event not necessarily have retry and attempt, I will manually construct it

invariant:
1. event log must be unique in conversationid, messageid, version
2. evenid is unique and hash of(cid,mid,version)
3. retry, and attempt is constructed from log
4. event log must be incremental under conversationid, messageid
5. conversation is done when it is exist in conversation table
6. no retry and attempt more than configuration
7. construct state is using graph I provided the state a-d as node and EVENT as edge


give me the working code as well as migration script
I want to use pg, kafka, and docker for worker for my whole system, use the same isolated network, use docker-compose to manage this, have a makefile, give the code insightful comments, code each chunk under 15 code complexity

in the end I want to run my code as make run, and down everything as make down

in conversation table: eventid must be introduce to prevent duplicate