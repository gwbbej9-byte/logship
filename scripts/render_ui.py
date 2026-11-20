#!/usr/bin/env python3
"""Render a dark 920x520 terminal still of logship for README."""

from __future__ import annotations

from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "screenshots" / "ui.png"
W, H = 920, 520

BG = (13, 17, 23)
FG = (201, 209, 217)
DIM = (110, 118, 129)
CYAN = (88, 166, 255)
GREEN = (63, 185, 80)
YELLOW = (210, 153, 34)
RED = (248, 81, 73)
MUTED = (139, 148, 158)
BAR = (22, 27, 34)
RULE = (48, 54, 61)
ACCENT = (31, 111, 235)

ROWS = [
    (MUTED, " $ ", FG, "logship tail samples/app.log samples/api.log --listen :8080"),
    (None, None, None, None),  # spacer
    (CYAN, "[app.log] ", GREEN, "INFO  2026-11-17T11:02:01Z INFO started worker pool size=8"),
    (CYAN, "[api.log] ", GREEN, 'INFO  {"level":"INFO","msg":"listen :8080","svc":"api"}'),
    (CYAN, "[app.log] ", DIM, "DEBUG 2026-11-17T11:02:01Z DEBUG binding listener :8443"),
    (CYAN, "[api.log] ", DIM, 'DEBUG {"level":"DEBUG","msg":"GET /v1/users 200 12ms"}'),
    (CYAN, "[app.log] ", GREEN, 'INFO  {"ts":"...","level":"info","msg":"loaded config"}'),
    (CYAN, "[app.log] ", YELLOW, "WARN  2026-11-17T11:02:08Z WARN retry queue depth=41"),
    (CYAN, "[api.log] ", YELLOW, 'WARN  {"level":"WARN","msg":"rate limit approaching remaining=9"}'),
    (CYAN, "[app.log] ", RED, "ERROR 2026-11-17T11:02:11Z ERROR failed to flush wal: no space left"),
    (CYAN, "[api.log] ", RED, 'ERROR {"level":"ERROR","msg":"db timeout after 3s query=select orders"}'),
    (CYAN, "[app.log] ", DIM, 'DEBUG {"level":"debug","msg":"gc cycle 12 freed=4.1MiB"}'),
    (CYAN, "[api.log] ", GREEN, 'INFO  {"level":"INFO","msg":"GET /healthz 200 1ms"}'),
    (CYAN, "[app.log] ", GREEN, "INFO  request id=a31c method=POST path=/ingest status=204"),
    (CYAN, "[api.log] ", RED, "ERROR 2026-11-17T11:02:20Z ERROR upstream 502 from billing"),
    (CYAN, "[app.log] ", YELLOW, "WARN  disk usage 87% on /var/lib/logship"),
    (CYAN, "[api.log] ", GREEN, "INFO  recovered connection pool=12"),
]


def load_font(size: int) -> ImageFont.FreeTypeFont | ImageFont.ImageFont:
    candidates = [
        Path(r"C:\Windows\Fonts\cascadiamono.ttf"),
        Path(r"C:\Windows\Fonts\CascadiaMono.ttf"),
        Path(r"C:\Windows\Fonts\consola.ttf"),
        Path(r"C:\Windows\Fonts\Cour.ttf"),
        Path(r"C:\Windows\Fonts\lucon.ttf"),
        Path("/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf"),
        Path("/System/Library/Fonts/Menlo.ttc"),
    ]
    for p in candidates:
        if p.exists():
            try:
                return ImageFont.truetype(str(p), size=size)
            except OSError:
                continue
    return ImageFont.load_default()


def main() -> None:
    OUT.parent.mkdir(parents=True, exist_ok=True)
    img = Image.new("RGB", (W, H), BG)
    draw = ImageDraw.Draw(img)
    font = load_font(16)
    title = load_font(15)
    small = load_font(14)

    # title bar
    draw.rectangle((0, 0, W, 36), fill=BAR)
    for i, c in enumerate(((255, 95, 86), (255, 189, 46), (39, 201, 63))):
        x = 16 + i * 18
        draw.ellipse((x, 12, x + 12, 24), fill=c)
    draw.text((88, 9), "logship  —  concurrent tail", font=title, fill=FG)

    y = 52
    x0 = 22
    for tag_color, tag, body_color, body in ROWS:
        if tag is None:
            y += 8
            continue
        draw.text((x0, y), tag, font=font, fill=tag_color)
        bbox = draw.textbbox((x0, y), tag, font=font)
        draw.text((bbox[2], y), body, font=font, fill=body_color)
        y += 24

    # footer
    fy = H - 52
    draw.line((0, fy, W, fy), fill=RULE, width=1)
    draw.rectangle((0, fy, W, H), fill=BAR)
    pad = 18
    pill = "--listen :8080"
    pb = draw.textbbox((0, 0), pill, font=small)
    pw, ph = pb[2] - pb[0] + 16, pb[3] - pb[1] + 10
    px, py = pad, fy + (52 - ph) // 2
    draw.rounded_rectangle((px, py, px + pw, py + ph), radius=6, fill=ACCENT)
    draw.text((px + 8, py + 4), pill, font=small, fill=(255, 255, 255))
    draw.text(
        (px + pw + 16, py + 4),
        "GET /healthz    GET /lines?n=50    files=2    following",
        font=small,
        fill=MUTED,
    )

    img.save(OUT, "PNG")
    print(f"wrote {OUT} {img.size[0]}x{img.size[1]}")


if __name__ == "__main__":
    main()
