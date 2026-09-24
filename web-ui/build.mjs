import { build } from 'esbuild';
import { readFile, writeFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const directory = path.dirname(fileURLToPath(import.meta.url));
const output = path.resolve(directory, '../internal/web/components.min.js');
const result = await build({
  absWorkingDir: directory,
  entryPoints: ['components.js'],
  outfile: output,
  bundle: true,
  minify: true,
  format: 'iife',
  target: 'es2022',
  legalComments: 'inline',
  metafile: true,
  write: false,
});

// Include the licenses of the runtime packages in the embedded JS as well as
// in a readable file. Go builds consume the checked-in bundle directly.
const packages = new Set(Object.keys(result.metafile.inputs).flatMap(input => {
  const match = input.match(/node_modules\/((?:@[^/]+\/)?[^/]+)/);
  return match ? [match[1]] : [];
}));
const notices = [];
for (const name of [...packages].sort()) {
  const root = path.join(directory, 'node_modules', name);
  const pkg = JSON.parse(await readFile(path.join(root, 'package.json'), 'utf8'));
  let license;
  for (const filename of ['LICENSE', 'LICENSE.md', 'LICENSE.txt', 'license', 'LICENSE-MIT']) {
    try { license = await readFile(path.join(root, filename), 'utf8'); break; }
    catch (error) { if (error.code !== 'ENOENT') throw error; }
  }
  if (!license) throw new Error(`Missing license for ${name}`);
  notices.push(`${name} ${pkg.version}\n${license.trim()}`);
}
const licenses = notices.join('\n\n');
for (const file of result.outputFiles) {
  const prefix = file.path.endsWith('.js') ? `/*!\n${licenses.replaceAll('*/', '* /')}\n*/\n` : '';
  await writeFile(file.path, prefix + file.text);
}
await writeFile(path.resolve(directory, '../internal/web/components.LICENSE.txt'), licenses + '\n');
