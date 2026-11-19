import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const root = join(dirname(fileURLToPath(import.meta.url)), '..');
const script = join(root, 'hooks', 'voice.mjs');
const full = readFileSync(join(root, 'voice', 'full.md'), 'utf8');
const lite = readFileSync(join(root, 'voice', 'lite.md'), 'utf8');

function hook(payload, env = {}) {
  const input = typeof payload === 'string' ? payload : JSON.stringify(payload);
  const { ALT_VOICE_DEFAULT, CLAUDE_PLUGIN_DATA, ...base } = process.env;
  const r = spawnSync(process.execPath, [script], { input, env: { ...base, CLAUDE_PLUGIN_ROOT: root, ...env }, encoding: 'utf8' });
  assert.equal(r.status, 0);
  return r.stdout;
}

function context(out) {
  return JSON.parse(out).hookSpecificOutput.additionalContext;
}

const dataEnv = () => ({ CLAUDE_PLUGIN_DATA: mkdtempSync(join(tmpdir(), 'alt-voice-test-')) });
const start = (id, event = 'SessionStart') => ({ session_id: id, hook_event_name: event });
const say = (id, prompt) => ({ session_id: id, hook_event_name: 'UserPromptSubmit', prompt });

test('SessionStart with no mode injects full', () => {
  const out = hook(start('s1'), dataEnv());
  assert.equal(JSON.parse(out).hookSpecificOutput.hookEventName, 'SessionStart');
  assert.equal(context(out), full);
});

test('ALT_VOICE_DEFAULT overrides the default', () => {
  assert.equal(context(hook(start('s1'), { ...dataEnv(), ALT_VOICE_DEFAULT: 'lite' })), lite);
});

test('switch to lite writes the mode and SubagentStart injects lite', () => {
  const env = dataEnv();
  assert.equal(context(hook(say('s1', '  /alt-voice lite  '), env)), 'Voice mode: lite');
  assert.equal(readFileSync(join(env.CLAUDE_PLUGIN_DATA, 'modes', 's1'), 'utf8'), 'lite');
  const out = hook(start('s1', 'SubagentStart'), env);
  assert.equal(JSON.parse(out).hookSpecificOutput.hookEventName, 'SubagentStart');
  assert.equal(context(out), lite);
});

test('off stops injection, by command or phrase', () => {
  for (const phrase of ['/alt-voice off', 'stop voice', 'Normal mode']) {
    const env = dataEnv();
    assert.equal(context(hook(say('s1', phrase), env)), 'Voice mode: off');
    assert.equal(hook(start('s1'), env), '');
  }
});

test('an ordinary prompt gives no output', () => {
  assert.equal(hook(say('s1', 'please run /alt-voice lite later'), dataEnv()), '');
});

test('garbage stdin exits 0 with no output', () => {
  assert.equal(hook('not json {', dataEnv()), '');
  assert.equal(hook('', dataEnv()), '');
});

test('unknown event gives no output', () => {
  assert.equal(hook(start('s1', 'PreToolUse'), dataEnv()), '');
});

test('unsafe session id gives no output', () => {
  assert.equal(hook(say('../x', '/alt-voice lite'), dataEnv()), '');
});

test('two sessions keep independent modes', () => {
  const env = dataEnv();
  hook(say('a', '/alt-voice lite'), env);
  hook(say('b', '/alt-voice off'), env);
  assert.equal(context(hook(start('a'), env)), lite);
  assert.equal(hook(start('b'), env), '');
  assert.equal(context(hook(start('c'), env)), full);
});

test('missing CLAUDE_PLUGIN_DATA falls back to the temp dir', (t) => {
  const id = `fallback-${process.pid}-${Date.now()}`;
  t.after(() => rmSync(join(tmpdir(), 'alt-voice', 'modes', id), { force: true }));
  assert.equal(context(hook(say(id, '/alt-voice lite'))), 'Voice mode: lite');
  assert.equal(readFileSync(join(tmpdir(), 'alt-voice', 'modes', id), 'utf8'), 'lite');
  assert.equal(context(hook(start(id))), lite);
});
