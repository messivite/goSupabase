import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'goSupaBase',
  description: 'Build Go APIs fast with a Supabase backend',
  base: '/goSupabase/',

  head: [
    ['link', { rel: 'icon', href: '/goSupabase/favicon.ico' }],
  ],

  themeConfig: {
    logo: 'https://cdn.simpleicons.org/supabase/3ECF8E',

    nav: [
      { text: 'Guide', link: '/guide/introduction' },
      { text: 'Reference', link: '/reference/cli' },
      { text: 'Advanced', link: '/advanced/auth' },
      {
        text: 'Links',
        items: [
          { text: 'GitHub', link: 'https://github.com/messivite/goSupabase' },
          {
            text: 'LinkedIn — Mustafa Aksoy',
            link: 'https://www.linkedin.com/in/mustafa-aksoy-87532a385/',
          },
          { text: 'npm profile', link: 'https://www.npmjs.com/~mustafaaksoy41' },
          { text: 'Go Docs', link: 'https://pkg.go.dev/github.com/messivite/gosupabase' },
        ],
      },
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Common questions',
          items: [
            {
              text: 'How do I verify Supabase JWTs and JWKS?',
              link: '/advanced/auth#how-do-i-verify-supabase-jwts-and-jwks',
            },
          ],
        },
        {
          text: 'Getting Started',
          items: [
            { text: 'Introduction', link: '/guide/introduction' },
            { text: 'Installation', link: '/guide/installation' },
            { text: 'Quick Start', link: '/guide/quick-start' },
          ],
        },
        {
          text: 'Usage',
          items: [
            { text: 'Developer Flows', link: '/guide/developer-flows' },
            { text: 'Setup Wizard', link: '/guide/setup' },
            { text: 'Configuration', link: '/guide/configuration' },
          ],
        },
      ],
      '/reference/': [
        {
          text: 'Common questions',
          items: [
            {
              text: 'How do I verify Supabase JWTs and JWKS?',
              link: '/advanced/auth#how-do-i-verify-supabase-jwts-and-jwks',
            },
          ],
        },
        {
          text: 'Reference',
          items: [
            { text: 'CLI Commands', link: '/reference/cli' },
            { text: 'YAML Schema', link: '/reference/yaml-schema' },
            { text: 'Environment Variables', link: '/reference/environment-variables' },
            { text: 'Project Structure', link: '/reference/project-structure' },
            { text: 'Postman collection', link: '/reference/postman' },
          ],
        },
      ],
      '/advanced/': [
        {
          text: 'Advanced',
          items: [
            {
              text: 'How do I verify Supabase JWTs and JWKS?',
              link: '/advanced/auth#how-do-i-verify-supabase-jwts-and-jwks',
            },
            { text: 'Auth & JWT (full guide)', link: '/advanced/auth' },
            { text: 'Deployment', link: '/advanced/deployment' },
            { text: 'CI/CD & Releases', link: '/advanced/ci-cd' },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/messivite/goSupabase' },
      {
        icon: {
          svg: '<svg role="img" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" aria-hidden="true"><title>npm</title><path fill="currentColor" d="M1.763 0C.786 0 0 .786 0 1.763v20.474C0 23.214.786 24 1.763 24h20.474c.977 0 1.763-.786 1.763-1.763V1.763C24 .786 23.214 0 22.237 0zM5.13 5.323l13.837.019-.009 13.834h-3.464l.01-10.382h-3.456L12.04 19.17H5.113z"/></svg>',
        },
        link: 'https://www.npmjs.com/~mustafaaksoy41',
        ariaLabel: 'npm profile',
      },
      {
        icon: {
          svg: '<svg role="img" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" aria-hidden="true"><title>LinkedIn</title><path fill="currentColor" d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/></svg>',
        },
        link: 'https://www.linkedin.com/in/mustafa-aksoy-87532a385/',
        ariaLabel: 'Mustafa Aksoy on LinkedIn',
      },
    ],

    search: {
      provider: 'local',
    },

    editLink: {
      pattern: 'https://github.com/messivite/goSupabase/edit/main/docs/:path',
      text: 'Edit this page on GitHub',
    },

    footer: {
      message:
        'Released under the MIT License. · Made with <a href="https://www.linkedin.com/in/mustafa-aksoy-87532a385/" target="_blank" rel="noopener noreferrer">Mustafa Aksoy</a>',
      copyright: 'Copyright 2024-present goSupaBase contributors',
    },
  },
})
