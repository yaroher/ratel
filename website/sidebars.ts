import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    {
      type: 'category',
      label: 'Getting Started',
      collapsed: false,
      items: [
        'getting-started/installation',
        'getting-started/quick-start-proto',
        'getting-started/quick-start-sql',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/defining-tables',
        'guides/column-types',
        'guides/constraints',
        'guides/relations',
        'guides/schema-generation',
        'guides/query-building',
        'guides/executing-queries',
        'guides/repository-pattern',
        'guides/migrations',
      ],
    },
    {
      type: 'category',
      label: 'API Reference',
      items: [
        'api/proto-options',
        'api/ddl',
        'api/dml',
        'api/schema',
        'api/exec',
        'api/repository',
      ],
    },
    {
      type: 'category',
      label: 'CLI Reference',
      items: [
        'cli/ratel-schema',
        'cli/ratel-diff',
        'cli/ratel-generate',
        'cli/protoc-gen-ratel',
      ],
    },
    {
      type: 'category',
      label: 'Examples',
      items: [
        'examples/store',
        'examples/common-patterns',
      ],
    },
  ],
};

export default sidebars;
