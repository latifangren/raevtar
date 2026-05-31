---
version: alpha
name: Raevtar
description: Personal platform — blog, dashboard, API. Dark-first, developer aesthetic.
colors:
  primary: "#0F172A"      # Slate 900 — background utama
  secondary: "#1E293B"    # Slate 800 — card/surface
  accent: "#22C55E"       # Green 500 — aksen utama (status online, link, CTA)
  accent-hover: "#16A34A" # Green 600
  text-primary: "#F1F5F9" # Slate 100 — teks utama
  text-secondary: "#94A3B8" # Slate 400 — teks redup
  border: "#334155"       # Slate 700 — border
  danger: "#EF4444"       # Red 500 — offline status, error
  warning: "#F59E0B"      # Amber 500 — warning

typography:
  h1:
    fontFamily: "'Inter', system-ui, sans-serif"
    fontSize: 2.25rem
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: "-0.02em"
  h2:
    fontFamily: "'Inter', system-ui, sans-serif"
    fontSize: 1.5rem
    fontWeight: 600
    lineHeight: 1.2
  body:
    fontFamily: "'Inter', system-ui, sans-serif"
    fontSize: 1rem
    fontWeight: 400
    lineHeight: 1.6
  mono:
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace"
    fontSize: 0.875rem
    fontWeight: 400

rounded:
  sm: 4px
  md: 8px
  lg: 12px
  xl: 16px

spacing:
  xs: 4px
  sm: 8px
  md: 16px
  lg: 24px
  xl: 32px
  2xl: 48px

components:
  card:
    backgroundColor: "{colors.secondary}"
    rounded: "{rounded.lg}"
    padding: 16px
    border: "1px solid {colors.border}"
  card-hover:
    backgroundColor: "#243447"
    border: "1px solid {colors.accent}"
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "#000000"
    rounded: "{rounded.md}"
    padding: "8px 16px"
    fontWeight: 600
  button-primary-hover:
    backgroundColor: "{colors.accent-hover}"
  badge:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.text-secondary}"
    rounded: "{rounded.sm}"
    padding: "2px 8px"
    fontSize: "0.75rem"
  status-online:
    backgroundColor: "{colors.accent}"
    rounded: "50%"
    size: 8px
  status-offline:
    backgroundColor: "{colors.danger}"
    rounded: "50%"
    size: 8px
  nav-link:
    textColor: "{colors.text-secondary}"
    fontWeight: 500
  nav-link-active:
    textColor: "{colors.accent}"
    border: "0 0 2px 0 solid {colors.accent}"
---

# Raevtar Design System

## Overview

Dark-first, developer-centric design. Terminal-inspired but polished — bukan "hacker green on black" stereotype. Warna hijau sebagai aksen tunggal biar gak berisik.

## Colors

- **Primary (#0F172A):** Slate gelap sebagai kanvas. Matanya gak sakit.
- **Accent (#22C55E):** Satu-satunya warna cerah. Buat link, status online, tombol.
- **Danger (#EF4444):** Khusus buat status server offline.

Gak ada gradien, gak ada warna ramai. Dua warna abu, satu hijau, satu merah. Selesai.

## Typography

Inter untuk semua teks. JetBrains Mono untuk kode. Font-size scale mengikuti Tailwind default.

## Components

Card adalah unit dasar. Semua konten di dalam card — termasuk blog post dan server card.

## Do's and Don'ts

- ✅ Gunakan hijau untuk satu elemen interaktif per halaman
- ✅ Card konsisten: dark surface, border tipis
- ❌ Jangan pakai gradient
- ❌ Jangan lebih dari satu warna aksen
