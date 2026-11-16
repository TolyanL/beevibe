from __future__ import annotations

import os
import shutil
import subprocess
from pathlib import Path

from PIL import Image, ImageOps
from pptx import Presentation
from pptx.dml.color import RGBColor
from pptx.enum.dml import MSO_LINE_DASH_STYLE
from pptx.enum.shapes import MSO_AUTO_SHAPE_TYPE as SHAPE
from pptx.enum.text import MSO_ANCHOR, PP_ALIGN
from pptx.oxml.xmlchemy import OxmlElement
from pptx.util import Inches, Pt

ROOT = Path(__file__).resolve().parents[1]
ASSETS = ROOT / "assets"
SHOTS = ASSETS / "shots"
FONTS = ASSETS / "fonts"
OUT = ROOT / "beevibe.pptx"
BUILD = ROOT / "build"

W, H = 13.333, 7.5
CREAM = "F3F0E6"
PAPER = "FFFDF7"
INK = "141414"
NEAR = "0D0D0B"
MUTED = "706B5E"
FAINT = "98917F"
YEL = "FFCC00"
YEL_DARK = "E6B800"
GREEN = "22C55E"
LIME = "84CC16"
YELLOW = "EAB308"
ORANGE = "F97316"
RED = "EF4444"
WHITE = "FFFFFF"

DISPLAY = "Unbounded"
BODY = "Manrope"
MONO = "JetBrains Mono"

SHOT_FILES = {i: Path(f"/home/dino/Downloads/Telegram Desktop/{i}.png") for i in range(1, 7)}


def rgb(hex_color: str) -> RGBColor:
    return RGBColor.from_string(hex_color)


def set_transparency(fill, transparency: float):
    if transparency is None:
        return
    solid = fill._xPr.solidFill
    if solid is None:
        return
    for child in list(solid):
        if child.tag.endswith('}srgbClr'):
            alpha = OxmlElement('a:alpha')
            alpha.set('val', str(int((1 - transparency) * 100000)))
            child.append(alpha)
            return


def rect(slide, x, y, w, h, fill=CREAM, line=None, radius=True, line_width=0.8, transparency=None):
    shape_type = SHAPE.ROUNDED_RECTANGLE if radius else SHAPE.RECTANGLE
    shp = slide.shapes.add_shape(shape_type, Inches(x), Inches(y), Inches(w), Inches(h))
    shp.fill.solid()
    shp.fill.fore_color.rgb = rgb(fill)
    set_transparency(shp.fill, transparency)
    if line:
        shp.line.color.rgb = rgb(line)
        shp.line.width = Pt(line_width)
    else:
        shp.line.fill.background()
    return shp


def ellipse(slide, x, y, w, h, fill=YEL, line=None, line_width=0.8, transparency=None):
    shp = slide.shapes.add_shape(SHAPE.OVAL, Inches(x), Inches(y), Inches(w), Inches(h))
    shp.fill.solid()
    shp.fill.fore_color.rgb = rgb(fill)
    set_transparency(shp.fill, transparency)
    if line:
        shp.line.color.rgb = rgb(line)
        shp.line.width = Pt(line_width)
    else:
        shp.line.fill.background()
    return shp


def text(
    slide,
    txt,
    x,
    y,
    w,
    h,
    *,
    size=18,
    font=BODY,
    bold=False,
    color=INK,
    align=PP_ALIGN.LEFT,
    valign=MSO_ANCHOR.TOP,
    margin=0,
    spacing=1.0,
    italic=False,
):
    box = slide.shapes.add_textbox(Inches(x), Inches(y), Inches(w), Inches(h))
    tf = box.text_frame
    tf.clear()
    tf.word_wrap = True
    tf.margin_left = Inches(margin)
    tf.margin_right = Inches(margin)
    tf.margin_top = Inches(margin)
    tf.margin_bottom = Inches(margin)
    tf.vertical_anchor = valign
    lines = txt.split("\n")
    for idx, line in enumerate(lines):
        p = tf.paragraphs[0] if idx == 0 else tf.add_paragraph()
        p.alignment = align
        p.line_spacing = spacing
        p.space_after = Pt(0)
        r = p.add_run()
        r.text = line
        r.font.name = font
        r.font.size = Pt(size)
        r.font.bold = bold
        r.font.italic = italic
        r.font.color.rgb = rgb(color)
    return box


def round_label(slide, txt, x, y, w, fill=YEL, color=INK, size=10, h=0.31, font=MONO):
    rect(slide, x, y, w, h, fill=fill, radius=True)
    text(
        slide,
        txt,
        x + 0.08,
        y + 0.04,
        w - 0.16,
        h - 0.05,
        size=size,
        font=font,
        bold=True,
        color=color,
        align=PP_ALIGN.CENTER,
        valign=MSO_ANCHOR.MIDDLE,
    )


def dots(slide, x, y, w, h, color=FAINT, transparency=0.72):
    shp = rect(slide, x, y, w, h, fill=color, radius=False, transparency=transparency)
    return shp


def base_slide(prs, dark=False, number=None, section=None):
    slide = prs.slides.add_slide(prs.slide_layouts[6])
    bg = NEAR if dark else CREAM
    rect(slide, 0, 0, W, H, fill=bg, radius=False)
    if not dark:
        ellipse(slide, -1.4, -1.0, 4.2, 3.0, fill=YEL, transparency=0.92)
        ellipse(slide, 10.9, 0.1, 3.3, 2.5, fill=YEL, transparency=0.95)
        dots(slide, 0, 0, W, H, color=INK, transparency=0.90)
    else:
        ellipse(slide, 10.6, -1.6, 4.5, 4.5, fill=YEL, transparency=0.90)
    if number is not None:
        text(
            slide,
            f"{number:02d}",
            12.45,
            7.02,
            0.48,
            0.22,
            size=10,
            font=MONO,
            bold=True,
            color=MUTED if not dark else FAINT,
            align=PP_ALIGN.RIGHT,
        )
    if section:
        text(
            slide,
            section.upper(),
            0.56,
            0.32,
            5.8,
            0.24,
            size=9,
            font=MONO,
            bold=True,
            color=MUTED if not dark else FAINT,
        )
    return slide


def header(slide, title, subtitle=None, dark=False, kicker=None, title_size=31):
    if kicker:
        round_label(slide, kicker, 0.55, 0.70, 1.30, fill=YEL, color=INK, size=9)
        y = 1.02
    else:
        y = 0.72
    text(slide, title, 0.55, y, 12.1, 0.82, size=title_size, font=DISPLAY, bold=True, color=WHITE if dark else INK)
    if subtitle:
        text(slide, subtitle, 0.57, y + 0.85, 11.9, 0.52, size=14.5, font=BODY, color="CFCBB8" if dark else MUTED)


def add_shot(slide, path: Path, x, y, w, h, label=None, crop="center", dark_bg=True, shadow=True):
    frame_fill = NEAR if dark_bg else "DAD4C7"
    if shadow:
        rect(slide, x + 0.08, y + 0.12, w, h, fill=frame_fill, radius=True, transparency=0.25)
    rect(slide, x, y, w, h, fill=frame_fill, radius=True)
    dots(slide, x + 0.18, y + 0.19, 0.10, 0.10, color=RED, transparency=0.05)
    dots(slide, x + 0.40, y + 0.19, 0.10, 0.10, color=YELLOW, transparency=0.05)
    dots(slide, x + 0.62, y + 0.19, 0.10, 0.10, color=GREEN, transparency=0.05)
    inner_x, inner_y = x + 0.10, y + 0.42
    inner_w, inner_h = w - 0.20, h - 0.53
    img = Image.open(path).convert("RGB")
    if crop == "top":
        target_ratio = inner_w / inner_h
        crop_h = int(img.width / target_ratio)
        img = img.crop((0, 0, img.width, min(crop_h, img.height)))
    else:
        img = ImageOps.fit(img, (1800, int(1800 * inner_h / inner_w)), method=Image.Resampling.LANCZOS)
    tmp = SHOTS / f"_fit_{path.stem}_{int(inner_w*100)}x{int(inner_h*100)}.jpg"
    img.save(tmp, quality=88)
    slide.shapes.add_picture(str(tmp), Inches(inner_x), Inches(inner_y), width=Inches(inner_w), height=Inches(inner_h))
    if label:
        round_label(slide, label, x + 0.16, y + h - 0.28, min(2.7, w - 0.32), fill=YEL, color=INK, size=8.5, h=0.25)


def arrow(slide, x, y, w=0.42, color=YEL_DARK, vertical=False):
    shp = slide.shapes.add_shape(
        SHAPE.DOWN_ARROW if vertical else SHAPE.RIGHT_ARROW, Inches(x), Inches(y), Inches(w), Inches(w)
    )
    shp.fill.solid()
    shp.fill.fore_color.rgb = rgb(color)
    shp.line.fill.background()
    return shp


def stat_card(slide, x, y, w, h, value, label, fill=PAPER, accent=YEL, dark=False):
    rect(slide, x, y, w, h, fill=fill, line=None if dark else "DED8C8", radius=True)
    ellipse(slide, x + 0.18, y + 0.18, 0.26, 0.26, fill=accent)
    text(slide, value, x + 0.18, y + 0.56, w - 0.36, 0.72, size=31, font=DISPLAY, bold=True, color=INK)
    text(slide, label, x + 0.2, y + h - 0.58, w - 0.4, 0.5, size=11.5, font=BODY, color=MUTED)


def bullet(slide, x, y, w, txt, *, size=14, dot=YEL, color=INK, font=BODY):
    ellipse(slide, x, y + 0.11, 0.13, 0.13, fill=dot)
    text(slide, txt, x + 0.28, y, w - 0.28, 0.48, size=size, font=font, color=color)


def prepare_shots():
    SHOTS.mkdir(parents=True, exist_ok=True)
    crops = {
        1: ("landing", (0, 120, 1920, 1900)),
        2: ("generator", (0, 0, 1920, 1500)),
        3: ("catalog", (0, 0, 1920, 1450)),
        4: ("profile", (0, 0, 1920, 1500)),
        5: ("workspace", (0, 0, 1920, 1500)),
        6: ("public", (0, 0, 1920, 1500)),
    }
    result = {}
    for n, (name, box) in crops.items():
        out = SHOTS / f"{n}-{name}.png"
        with Image.open(SHOT_FILES[n]) as im:
            im.crop(box).save(out)
        result[n] = out
    return result


def build():
    SHOTS.mkdir(parents=True, exist_ok=True)
    shots = prepare_shots()
    prs = Presentation()
    prs.slide_width = Inches(W)
    prs.slide_height = Inches(H)
    prs.core_properties.title = "BeeVibe — игры из одной фразы"
    prs.core_properties.subject = "Презентация продукта и фронтенда BeeVibe"
    prs.core_properties.author = "BeeVibe"

    # 1. Cover
    s = base_slide(prs, dark=True, number=1, section="BeeVibe / product deck")
    text(s, "BeeVibe", 0.62, 0.78, 7.4, 0.9, size=46, font=DISPLAY, bold=True, color=WHITE)
    text(s, "игры из одной фразы", 0.65, 1.75, 7.4, 0.72, size=28, font=DISPLAY, bold=True, color=YEL)
    text(
        s,
        "Продукт, который превращает короткий промпт в готовую HTML5-игру,\nкаталог, публичную ссылку и живой рабочий процесс владельца.",
        0.66,
        2.66,
        6.5,
        1.2,
        size=16.5,
        font=BODY,
        color="D9D5C5",
        spacing=1.25,
    )
    round_label(s, "AI GAMEGEN · FRONTEND + API", 0.66, 4.08, 2.75, fill=YEL, color=INK, size=10)
    add_shot(s, shots[1], 7.15, 1.12, 5.55, 4.95, label="лендинг / генератор", dark_bg=True)
    text(s, "Одна идея → playable build", 0.68, 6.48, 5.4, 0.35, size=12, font=MONO, color=FAINT)

    # 2. Problem
    s = base_slide(prs, number=2, section="Problem")
    header(
        s,
        "Идея есть. Игры — пока нет.",
        "Обычный путь от задумки до играбельного прототипа слишком дорогой и медленный.",
        kicker="01 / PROBLEM",
    )
    stat_card(s, 0.60, 2.08, 3.75, 2.10, "24ч+", "от идеи до первого билда", fill=WHITE)
    stat_card(s, 4.60, 2.08, 3.75, 2.10, "8", "этапов, в которых легко потеряться", fill=WHITE, accent=ORANGE)
    stat_card(s, 8.60, 2.08, 3.90, 2.10, "0", "гарантии, что получится игра", fill=WHITE, accent=RED)
    rect(s, 0.6, 4.62, 11.9, 1.65, fill=INK, radius=True)
    text(s, "BeeVibe", 0.95, 4.93, 2.25, 0.5, size=23, font=DISPLAY, bold=True, color=YEL)
    text(
        s,
        "Фраза → готовый single-file HTML5-проект: без установки, без сборки и без ручной возни с кодом.",
        3.0,
        4.93,
        9.0,
        0.85,
        size=16,
        font=BODY,
        color=WHITE,
    )
    text(s, "За секунды, а не дни", 0.96, 5.60, 4.0, 0.32, size=11, font=MONO, color=FAINT)

    # 3. Pipeline
    s = base_slide(prs, number=3, section="How it works")
    header(
        s,
        "Один промпт — восемь этапов конвейера",
        "Фронтенд показывает прогресс в реальном времени через SSE. Каждый этап — отдельное событие, каждая версия — отдельная запись истории.",
        kicker="02 / PIPELINE",
    )
    stages = [
        ("QUEUED", "очередь"),
        ("ANALYZE", "идея"),
        ("PLAN", "правила"),
        ("CODE", "код"),
        ("ASSETS", "визуал"),
        ("REVIEW", "ревью"),
        ("ASSEMBLE", "сборка"),
        ("CHECK", "проверка"),
    ]
    x0, y0, bw, gap = 0.58, 2.35, 1.35, 0.17
    for i, (code, label) in enumerate(stages):
        x = x0 + i * (bw + gap)
        fill = YEL if i in (0, 7) else WHITE
        rect(s, x, y0, bw, 1.34, fill=fill, line=None if i in (0, 7) else "DED8C8", radius=True)
        text(s, f"{i+1:02d}", x + 0.12, y0 + 0.12, 0.45, 0.22, size=9, font=MONO, bold=True, color=MUTED)
        text(s, code, x + 0.11, y0 + 0.47, bw - 0.22, 0.42, size=11.5, font=DISPLAY, bold=True, color=INK)
        text(s, label, x + 0.11, y0 + 0.96, bw - 0.22, 0.25, size=9.5, font=BODY, color=MUTED)
        if i < len(stages) - 1:
            arrow(s, x + bw + 0.02, y0 + 0.60, w=0.12, color=YEL_DARK)
    rect(s, 0.6, 4.20, 5.65, 1.62, fill=WHITE, radius=True)
    text(s, "SSE events", 0.86, 4.45, 2.4, 0.34, size=15, font=DISPLAY, bold=True, color=INK)
    text(s, "stage · artifact · done · error", 0.86, 4.96, 4.7, 0.36, size=12.5, font=MONO, color=MUTED)
    text(s, "EventSource с автоматическим reconnect", 0.86, 5.33, 4.7, 0.30, size=11.5, font=BODY, color=MUTED)
    rect(s, 6.55, 4.20, 5.95, 1.62, fill=INK, radius=True)
    text(s, "Реальные артефакты", 6.82, 4.45, 3.1, 0.34, size=15, font=DISPLAY, bold=True, color=WHITE)
    text(s, "spec · config · versions · cost", 6.82, 4.96, 5.0, 0.36, size=12.5, font=MONO, color=YEL)
    text(s, "Ничего не выдумано — всё из backend API", 6.82, 5.33, 5.0, 0.30, size=11.5, font=BODY, color="C9C5B5")

    # 4. Generator
    s = base_slide(prs, number=4, section="Generator")
    header(
        s,
        "Генератор: идея → управляемая сборка",
        "Промпт, жанр, стиль, размер карты и скорость сборки. Один экран вместо настройки пайплайна вручную.",
        kicker="03 / GENERATOR",
    )
    add_shot(s, shots[2], 0.6, 2.06, 8.15, 4.65, label="screenshot 02 / generator", dark_bg=True)
    rect(s, 9.05, 2.06, 3.48, 4.65, fill=WHITE, radius=True)
    text(s, "Что задаёт пользователь", 9.31, 2.38, 2.96, 0.5, size=16, font=DISPLAY, bold=True, color=INK)
    bullet(s, 9.31, 3.04, 2.9, "промпт и название")
    bullet(s, 9.31, 3.50, 2.9, "жанр и визуальный стиль")
    bullet(s, 9.31, 3.96, 2.9, "размер карты и темп")
    bullet(s, 9.31, 4.42, 2.9, "стоимость и баланс")
    rect(s, 9.30, 5.05, 2.98, 1.18, fill=YEL, radius=True)
    text(s, "240K / 320K", 9.50, 5.24, 2.6, 0.42, size=22, font=DISPLAY, bold=True, color=INK)
    text(s, "быстрая / обычная сборка", 9.50, 5.74, 2.6, 0.27, size=10.5, font=BODY, color=INK)

    # 5. Catalog
    s = base_slide(prs, number=5, section="Catalog")
    header(
        s,
        "Каталог: игры, которые можно потрогать",
        "Поиск, жанры, сортировка по рейтингу, времени и свежести. Оценка 1–5 в карточке и публичная страница в один клик.",
        kicker="04 / CATALOG",
    )
    add_shot(s, shots[3], 0.6, 2.03, 8.15, 4.70, label="screenshot 03 / catalog", dark_bg=False)
    rect(s, 9.05, 2.03, 3.48, 4.70, fill=WHITE, radius=True)
    text(s, "В каталоге", 9.30, 2.36, 2.9, 0.42, size=17, font=DISPLAY, bold=True, color=INK)
    bullet(s, 9.30, 2.98, 2.95, "поиск по названию и промпту")
    bullet(s, 9.30, 3.44, 2.95, "фильтры по жанру")
    bullet(s, 9.30, 3.90, 2.95, "sort: rating / hours / fresh")
    bullet(s, 9.30, 4.36, 2.95, "реальные aggregate из API")
    rect(s, 9.30, 5.04, 2.98, 1.18, fill=INK, radius=True)
    text(s, "1–5", 9.52, 5.20, 2.5, 0.48, size=28, font=DISPLAY, bold=True, color=YEL)
    text(s, "честная шкала оценок", 9.52, 5.72, 2.5, 0.28, size=11, font=BODY, color=WHITE)

    # 6. Workspace
    s = base_slide(prs, dark=True, number=6, section="Owner workspace")
    header(
        s,
        "Рабочее пространство владельца",
        "Версии, правки, конфиг, откаты, починка и оценка — всё рядом с живым превью игры.",
        dark=True,
        kicker="05 / WORKSPACE",
    )
    add_shot(s, shots[5], 0.6, 2.10, 7.0, 4.60, label="screenshot 05 / workspace", dark_bg=True)
    items = [
        ("ПРАВКА", "изменить игру промптом", YEL),
        ("КОНФИГ", "tempo · palette · map · difficulty", LIME),
        ("ОТКАТ", "вернуть предыдущую версию", ORANGE),
        ("ПОЧИНКА", "починка по лимиту попыток", RED),
        ("ОЦЕНКА", "1–5 и распределение голосов", GREEN),
    ]
    for i, (title, desc, color) in enumerate(items):
        y = 2.10 + i * 0.92
        rect(s, 7.95, y, 4.55, 0.76, fill="1D1D19", line="3A3A32", radius=True)
        ellipse(s, 8.14, y + 0.20, 0.31, 0.31, fill=color)
        text(s, title, 8.62, y + 0.10, 1.65, 0.26, size=11.5, font=MONO, bold=True, color=WHITE)
        text(s, desc, 8.62, y + 0.40, 3.5, 0.28, size=11, font=BODY, color="B9B5A4")

    # 7. Public play
    s = base_slide(prs, number=7, section="Public play")
    header(
        s,
        "Публичная игра без установки",
        "Отдельная share-страница, широкий stage, fullscreen и изолированный iframe с backend-HTML игры.",
        kicker="06 / PLAY",
    )
    add_shot(s, shots[6], 0.6, 2.06, 8.10, 4.65, label="screenshot 06 / public play", dark_bg=True)
    rect(s, 9.0, 2.06, 3.53, 4.65, fill=WHITE, radius=True)
    text(s, "Что важно", 9.26, 2.37, 2.9, 0.42, size=17, font=DISPLAY, bold=True, color=INK)
    bullet(s, 9.26, 2.98, 2.98, "ссылка /play/:id")
    bullet(s, 9.26, 3.44, 2.98, "iframe с песочницей")
    bullet(s, 9.26, 3.90, 2.98, "счётчик времени в игре")
    bullet(s, 9.26, 4.36, 2.98, "метаданные из публичного API")
    rect(s, 9.26, 5.05, 2.97, 1.15, fill=YEL, radius=True)
    text(s, "sandbox", 9.48, 5.22, 2.5, 0.40, size=18, font=MONO, bold=True, color=INK)
    text(s, "без доступа к origin приложения", 9.48, 5.71, 2.5, 0.28, size=10.5, font=BODY, color=INK)

    # 8. Profile & economy
    s = base_slide(prs, number=8, section="Profile & tokens")
    header(
        s,
        "Профиль: экономика и история",
        "Баланс, пополнение, список игр и журнал токен-операций. Цены видны до запуска сборки.",
        kicker="07 / PROFILE",
    )
    add_shot(s, shots[4], 0.6, 2.06, 8.15, 4.65, label="screenshot 04 / profile", dark_bg=False)
    rect(s, 9.05, 2.06, 3.48, 4.65, fill=WHITE, radius=True)
    stat_card(s, 9.28, 2.32, 3.02, 1.28, "1M", "стартовый грант", fill=CREAM, accent=YEL)
    bullet(s, 9.30, 3.86, 2.95, "generate · patch · topup")
    bullet(s, 9.30, 4.32, 2.95, "refund при failed generation")
    bullet(s, 9.30, 4.78, 2.95, "история версий и расходов")
    rect(s, 9.28, 5.36, 3.02, 0.96, fill=INK, radius=True)
    text(s, "320K", 9.48, 5.53, 2.5, 0.38, size=22, font=DISPLAY, bold=True, color=YEL)
    text(s, "обычная генерация", 9.48, 5.98, 2.5, 0.26, size=10.5, font=BODY, color=WHITE)

    # 9. Rating scale
    s = base_slide(prs, dark=True, number=9, section="Rating")
    header(
        s,
        "Реальная шкала 1–5 — не косметика",
        "Рейтинг хранится в PostgreSQL, приходит в aggregate и используется каталогом для сортировки.",
        dark=True,
        kicker="08 / RATING",
    )
    tones = [
        (5, "Восторг", GREEN),
        (4, "Хорошо", LIME),
        (3, "Нормально", YELLOW),
        (2, "Слабо", ORANGE),
        (1, "Плохо", RED),
    ]
    for i, (score, label, color) in enumerate(tones):
        x = 0.65 + i * 2.44
        rect(s, x, 2.25, 2.18, 2.18, fill=color, radius=True)
        text(s, str(score), x + 0.15, 2.45, 1.88, 1.0, size=46, font=DISPLAY, bold=True, color=INK, align=PP_ALIGN.CENTER)
        text(s, label, x + 0.14, 3.62, 1.9, 0.45, size=12.5, font=BODY, bold=True, color=INK, align=PP_ALIGN.CENTER)
    rect(s, 0.65, 4.93, 6.25, 1.55, fill="1D1D19", line="3A3A32", radius=True)
    text(s, "aggregate", 0.93, 5.18, 2.0, 0.30, size=11, font=MONO, bold=True, color=YEL)
    text(s, "average · total · counts[5]", 0.93, 5.62, 5.4, 0.42, size=15.5, font=MONO, color=WHITE)
    rect(s, 7.20, 4.93, 5.28, 1.55, fill=YEL, radius=True)
    text(s, "sort=rating", 7.48, 5.18, 2.3, 0.30, size=11, font=MONO, bold=True, color=INK)
    text(s, "ORDER BY average DESC, total DESC", 7.48, 5.62, 4.6, 0.42, size=13.5, font=MONO, color=INK)

    # 10. Architecture
    s = base_slide(prs, number=10, section="Architecture")
    header(
        s,
        "Архитектура: тонкий UI, честный API",
        "Vue 3 + TypeScript + OpenVue + Tailwind говорят с Go backend через REST и SSE. Session cookie и proxy — без CORS.",
        kicker="09 / ARCHITECTURE",
    )
    layers = [
        (0.62, "FRONTEND", "Vue 3 · TS · Pinia\nVue Router 5\nOpenVue · Tailwind 4", YEL),
        (4.33, "BACKEND", "Go 1.27 · chi\nREST /api · SSE /stream\nsession cookie", WHITE),
        (8.04, "DATA", "PostgreSQL 16\nmigrations · outbox\nKafka / SSE events", WHITE),
    ]
    for x, title, body, fill in layers:
        rect(s, x, 2.28, 3.40, 2.42, fill=fill, line=None if fill == YEL else "DED8C8", radius=True)
        text(s, title, x + 0.22, 2.56, 2.9, 0.42, size=17, font=DISPLAY, bold=True, color=INK)
        text(s, body, x + 0.22, 3.14, 2.95, 1.2, size=12.5, font=MONO, color=INK, spacing=1.35)
    arrow(s, 4.08, 3.25, 0.18)
    arrow(s, 7.79, 3.25, 0.18)
    rect(s, 0.62, 5.10, 12.0, 1.48, fill=INK, radius=True)
    text(s, "SECURITY", 0.90, 5.36, 1.8, 0.30, size=11, font=MONO, bold=True, color=YEL)
    text(
        s,
        "HttpOnly SameSite=Lax session · sandboxed game iframe · same-origin /api and /g proxy · no CDN assets",
        2.40,
        5.36,
        9.8,
        0.55,
        size=14,
        font=BODY,
        color=WHITE,
    )

    # 11. Done & limits
    s = base_slide(prs, number=11, section="Status")
    header(
        s,
        "Что проверено — и что осталось за кадром",
        "Фронтенд и контракт API полностью связаны. Ограничения локального окружения не меняют архитектуру продукта.",
        kicker="10 / STATUS",
    )
    rect(s, 0.60, 2.10, 6.00, 4.55, fill=WHITE, radius=True)
    text(s, "ГОТОВО", 0.90, 2.42, 2.4, 0.42, size=18, font=DISPLAY, bold=True, color=INK)
    done = [
        "авторизация, профиль, каталог, оценка",
        "SSE-прогресс и live preview",
        "patch / config / rollback / repair",
        "публичная игра и playtime",
        "light/dark, responsive, fullscreen",
        "go build + go test + bun type-check + build",
    ]
    for i, item in enumerate(done):
        bullet(s, 0.92, 3.05 + i * 0.52, 5.25, item, size=12.5, dot=GREEN)
    rect(s, 6.82, 2.10, 5.80, 4.55, fill=INK, radius=True)
    text(s, "ОГРАНИЧЕНИЯ", 7.12, 2.42, 3.2, 0.42, size=18, font=DISPLAY, bold=True, color=WHITE)
    limits = [
        "Kafka unavailable locally → SSE events empty",
        "LLM key needed for real generation",
        "golangci-lint not installed in sandbox",
        "Docker daemon permission denied in this session",
        "no fake ratings: only real backend state",
    ]
    for i, item in enumerate(limits):
        bullet(s, 7.14, 3.05 + i * 0.60, 5.05, item, size=12.5, dot=ORANGE, color="D7D3C3")

    # 12. Final
    s = base_slide(prs, dark=True, number=12, section="Final")
    ellipse(s, 9.25, 1.15, 3.45, 3.45, fill=YEL, transparency=0.04)
    rect(s, 9.89, 1.78, 2.18, 2.18, fill=INK, radius=True)
    text(s, "B", 10.30, 2.02, 1.35, 1.45, size=68, font=DISPLAY, bold=True, color=YEL, align=PP_ALIGN.CENTER)
    text(s, "BeeVibe", 0.70, 1.18, 8.2, 1.0, size=54, font=DISPLAY, bold=True, color=WHITE)
    text(s, "Одна фраза. Одна игра. Весь путь — под контролем.", 0.75, 2.32, 8.7, 0.65, size=22, font=DISPLAY, bold=True, color=YEL)
    text(
        s,
        "Собрано как продукт: от генератора и каталога до workspace, публичной страницы, оценок 1–5 и реального API.",
        0.75,
        3.20,
        7.9,
        1.0,
        size=16,
        font=BODY,
        color="D4D0BE",
        spacing=1.25,
    )
    round_label(s, "github.com/TolyanL/beevibe/tree/frontend", 0.75, 4.58, 5.18, fill=YEL, color=INK, size=10.5, h=0.44)
    text(s, "Спасибо!", 0.75, 5.55, 4.0, 0.55, size=25, font=DISPLAY, bold=True, color=WHITE)
    text(s, "Вопросы →", 0.77, 6.12, 2.3, 0.30, size=12, font=MONO, color=FAINT)

    prs.save(OUT)

    pdf_dir = BUILD / "pdf"
    pdf_dir.mkdir(parents=True, exist_ok=True)
    env = os.environ.copy()
    conf = Path.home() / ".local/share/fonts/beevibe-fontconfig.conf"
    if conf.exists():
        env["FONTCONFIG_FILE"] = str(conf)
    result = subprocess.run(
        [
            "soffice",
            "--headless",
            "--convert-to",
            "pdf",
            "--outdir",
            str(pdf_dir),
            str(OUT),
        ],
        check=False,
        env=env,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    if result.returncode != 0:
        print("PDF export skipped:", result.stderr.strip() or result.stdout.strip())
        return
    print(f"PPTX: {OUT}")
    print(f"PDF:  {pdf_dir / (OUT.stem + '.pdf')}")


if __name__ == "__main__":
    build()
