import { definePreset } from '@openvue/themes'
import Aura from '@openvue/themes/aura'

/**
 * Beeline-yellow arcade theme for OpenVue.
 *
 * Yellow is the action colour; near-black surfaces make the playable stage
 * feel like a piece of hardware rather than another dashboard.
 */
export const BeeVibePreset = definePreset(Aura, {
  semantic: {
    primary: {
      50: '#fffbe6',
      100: '#fff4bf',
      200: '#ffea80',
      300: '#ffe14d',
      400: '#ffd61a',
      500: '#ffcc00',
      600: '#e6b800',
      700: '#bf9800',
      800: '#997a00',
      900: '#735c00',
      950: '#3d3100',
    },
    colorScheme: {
      light: {
        primary: {
          color: '{primary.500}',
          contrastColor: '#141414',
          hoverColor: '{primary.600}',
          activeColor: '{primary.700}',
        },
        highlight: {
          background: '{primary.500}',
          focusBackground: '{primary.600}',
          color: '#141414',
          focusColor: '#141414',
        },
        surface: {
          0: '#fffefa',
          50: '#f8f6ee',
          100: '#f0ede1',
          200: '#e4dfcf',
          300: '#d2cbb7',
          400: '#a9a18c',
          500: '#7b7464',
          600: '#595448',
          700: '#3d3a32',
          800: '#26241f',
          900: '#181713',
          950: '#0d0d0b',
        },
      },
      dark: {
        primary: {
          color: '{primary.500}',
          contrastColor: '#141414',
          hoverColor: '{primary.400}',
          activeColor: '{primary.300}',
        },
        highlight: {
          background: 'rgba(255, 204, 0, 0.28)',
          focusBackground: 'rgba(255, 204, 0, 0.38)',
          color: '{primary.200}',
          focusColor: '{primary.100}',
        },
        surface: {
          0: '#fbfaf4',
          50: '#f2f1e9',
          100: '#e1e0d3',
          200: '#c6c5b3',
          300: '#a2a18d',
          400: '#777767',
          500: '#555549',
          600: '#3b3b33',
          700: '#2e2e28',
          800: '#242420',
          900: '#1d1d19',
          950: '#171714',
        },
      },
    },
  },
})
