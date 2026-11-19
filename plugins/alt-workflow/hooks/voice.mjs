import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const MODES = new Set(['lite', 'full', 'off']);
const SESSION_ID = /^[A-Za-z0-9_-]{1,128}$/;
const root = process.env.CLAUDE_PLUGIN_ROOT || join(dirname(fileURLToPath(import.meta.url)), '..');
const modesDir = join(process.env.CLAUDE_PLUGIN_DATA || join(tmpdir(), 'alt-voice'), 'modes');

function emit(event, text) {
  process.stdout.write(JSON.stringify({ hookSpecificOutput: { hookEventName: event, additionalContext: text } }));
}

function requestedMode(prompt) {
  const p = prompt.trim().toLowerCase();
  if (p === 'stop voice' || p === 'normal mode') return 'off';
  return p.match(/^\/alt-voice\s+(lite|full|off)$/)?.[1];
}

function run(input) {
  const { session_id: id, hook_event_name: event, prompt } = JSON.parse(input);
  if (!SESSION_ID.test(id)) return;
  const modeFile = join(modesDir, id);

  if (event === 'SessionStart' || event === 'SubagentStart') {
    let mode = process.env.ALT_VOICE_DEFAULT || 'full';
    try {
      mode = readFileSync(modeFile, 'utf8').trim();
    } catch {}
    if (!MODES.has(mode) || mode === 'off') return;
    emit(event, readFileSync(join(root, 'voice', `${mode}.md`), 'utf8'));
    return;
  }

  if (event === 'UserPromptSubmit' && typeof prompt === 'string') {
    const mode = requestedMode(prompt);
    if (!mode) return;
    mkdirSync(modesDir, { recursive: true });
    writeFileSync(modeFile, mode);
    emit(event, `Voice mode: ${mode}`);
  }
}

let input = '';
process.stdin.setEncoding('utf8');
process.stdin.on('data', (c) => { input += c; });
process.stdin.on('end', () => {
  try {
    run(input);
  } catch {}
  process.exitCode = 0;
});
