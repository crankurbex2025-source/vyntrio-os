#!/usr/bin/env node
/**
 * Generate branded PWA PNG icons for Vyntrio (Block 11R.7).
 * Dark surface + warm accent mark aligned with vyntrio.tokens.css.
 */
import { deflateSync } from "node:zlib";
import { mkdirSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const OUT_DIR = join(__dirname, "../public/icons");

const BG = [6, 8, 13, 255];
const ACCENT = [232, 93, 43, 255];

function crc32(buf) {
  let crc = 0xffffffff;
  for (let i = 0; i < buf.length; i++) {
    crc ^= buf[i];
    for (let j = 0; j < 8; j++) {
      crc = (crc >>> 1) ^ (crc & 1 ? 0xedb88320 : 0);
    }
  }
  return (crc ^ 0xffffffff) >>> 0;
}

function pngChunk(type, data) {
  const typeBuf = Buffer.from(type, "ascii");
  const len = Buffer.alloc(4);
  len.writeUInt32BE(data.length);
  const body = Buffer.concat([typeBuf, data]);
  const crc = Buffer.alloc(4);
  crc.writeUInt32BE(crc32(body));
  return Buffer.concat([len, body, crc]);
}

function inRoundedRect(x, y, left, top, width, height, radius) {
  if (x < left || y < top || x >= left + width || y >= top + height) {
    return false;
  }
  const right = left + width - 1;
  const bottom = top + height - 1;
  const r = Math.min(radius, width / 2, height / 2);

  const corners = [
    [left + r, top + r],
    [right - r, top + r],
    [left + r, bottom - r],
    [right - r, bottom - r],
  ];

  if (x < left + r && y < top + r) {
    return (x - corners[0][0]) ** 2 + (y - corners[0][1]) ** 2 <= r * r;
  }
  if (x > right - r && y < top + r) {
    return (x - corners[1][0]) ** 2 + (y - corners[1][1]) ** 2 <= r * r;
  }
  if (x < left + r && y > bottom - r) {
    return (x - corners[2][0]) ** 2 + (y - corners[2][1]) ** 2 <= r * r;
  }
  if (x > right - r && y > bottom - r) {
    return (x - corners[3][0]) ** 2 + (y - corners[3][1]) ** 2 <= r * r;
  }
  return true;
}

function drawIcon(size) {
  const pixels = Buffer.alloc(size * size * 4);
  const pad = Math.round(size * 0.2);
  const markW = size - pad * 2;
  const markH = size - pad * 2;
  const radius = Math.round(size * 0.1);

  for (let y = 0; y < size; y++) {
    for (let x = 0; x < size; x++) {
      const i = (y * size + x) * 4;
      const color = inRoundedRect(x, y, pad, pad, markW, markH, radius) ? ACCENT : BG;
      pixels[i] = color[0];
      pixels[i + 1] = color[1];
      pixels[i + 2] = color[2];
      pixels[i + 3] = color[3];
    }
  }

  const stride = size * 4 + size;
  const raw = Buffer.alloc(stride * size);
  for (let y = 0; y < size; y++) {
    const rowStart = y * stride;
    raw[rowStart] = 0;
    pixels.copy(raw, rowStart + 1, y * size * 4, (y + 1) * size * 4);
  }

  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(size, 0);
  ihdr.writeUInt32BE(size, 4);
  ihdr[8] = 8;
  ihdr[9] = 6;
  ihdr[10] = 0;
  ihdr[11] = 0;
  ihdr[12] = 0;

  return Buffer.concat([
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]),
    pngChunk("IHDR", ihdr),
    pngChunk("IDAT", deflateSync(raw, { level: 9 })),
    pngChunk("IEND", Buffer.alloc(0)),
  ]);
}

mkdirSync(OUT_DIR, { recursive: true });
for (const size of [192, 512]) {
  writeFileSync(join(OUT_DIR, `icon-${size}.png`), drawIcon(size));
}
console.log("Wrote PWA icons to", OUT_DIR);
