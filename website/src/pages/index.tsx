import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useBaseUrl from '@docusaurus/useBaseUrl';
import styles from './index.module.css';

const features = [
  {
    icon: '\u26A1',
    title: 'Type Safety',
    desc: 'Column names and types checked at compile time. No magic strings. No runtime reflection.',
  },
  {
    icon: '\uD83D\uDD17',
    title: 'Relations',
    desc: 'HasMany, BelongsTo, ManyToMany with eager loading. Type-safe relation options per query.',
  },
  {
    icon: '\uD83D\uDCE6',
    title: 'Code Generation',
    desc: 'Generate from Protobuf or SQL schema. Full DDL, DML, scanners, and repositories.',
  },
  {
    icon: '\uD83D\uDC18',
    title: 'PostgreSQL Native',
    desc: 'Built on pgx. JSONB, arrays, partial indexes, schemas, RLS \u2014 first-class support.',
  },
  {
    icon: '\uD83D\uDD04',
    title: 'Migrations',
    desc: 'Atlas-powered schema diffing. Auto-generate migration files from model changes.',
  },
  {
    icon: '\uD83D\uDEE1\uFE0F',
    title: 'Zero Overhead',
    desc: 'Go generics, no interface{} boxing. Direct struct scanning without reflection.',
  },
];

export default function Home() {
  return (
    <Layout title="Fearless PostgreSQL ORM for Go" description="Type-safe, compile-time checked PostgreSQL ORM for Go">
      {/* Hero */}
      <section className={styles.hero}>
        <div className={styles.heroBadge}>v0.1 — early access</div>
        <img src={useBaseUrl('/img/ratel.png')} alt="Ratel" className={styles.heroMascot} />
        <div className={styles.heroLogo}>
          R<span className={styles.heroLogoAccent}>A</span>TEL
        </div>
        <p className={styles.heroTagline}>
          Fearless, type-safe PostgreSQL ORM for Go. Compile-time checked queries. Zero reflection. Maximum performance.
        </p>
        <div className={styles.heroButtons}>
          <Link className={styles.btnPrimary} to="/docs/getting-started/installation">
            Get Started
          </Link>
          <Link className={styles.btnSecondary} href="https://github.com/yaroher/ratel">
            GitHub &rarr;
          </Link>
        </div>
        <div className={styles.installLine}>
          <span className={styles.installLinePrompt}>$</span> go install github.com/yaroher/ratel/cmd/ratel@latest
        </div>

        <div className={styles.codePreview}>
          <div className={styles.codeHeader}>
            <div className={styles.codeDotR} />
            <div className={styles.codeDotY} />
            <div className={styles.codeDotG} />
            store.go
          </div>
          <div className={styles.codeBody}>
{`// Type-safe queries — compiler catches typos
users, err := Users.Query(ctx, db,
    Users.SelectAll().
        Where(Users.Email.Eq("john@example.com")).
        OrderBy(Users.CreatedAt.Desc()).
        Limit(10),
)`}
          </div>
        </div>
      </section>

      {/* Features */}
      <section className={styles.features}>
        <div className={styles.featuresTitle}>// Why Ratel</div>
        <div className={styles.featuresGrid}>
          {features.map((f) => (
            <div key={f.title} className={styles.feature}>
              <div className={styles.featureIcon}>{f.icon}</div>
              <div className={styles.featureTitle}>{f.title}</div>
              <p className={styles.featureDesc}>{f.desc}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Comparison */}
      <section className={styles.comparison}>
        <div className={styles.comparisonTitle}>// Raw SQL vs Ratel</div>
        <div className={styles.comparisonSub}>Write less. Catch more.</div>
        <div className={styles.compareCols}>
          <div className={styles.compareCol}>
            <div className={styles.compareLabel}>Raw SQL + pgx</div>
            <div className={styles.compareCode}>
{`rows, err := db.Query(ctx,
  "SELECT id, email, name
   FROM users
   WHERE is_active = $1", true)

// Manual scanning...
// No compile-time checks
// Typos found at runtime`}
            </div>
          </div>
          <div className={styles.compareColHighlight}>
            <div className={styles.compareLabelHighlight}>Ratel</div>
            <div className={styles.compareCodeHighlight}>
{`users, err := Users.Query(ctx, db,
  Users.SelectAll().
    Where(Users.IsActive.Eq(true)),
)

// Auto scanning
// Compile-time column checks
// Type-safe predicates`}
            </div>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className={styles.cta}>
        <h2 className={styles.ctaTitle}>HONEY BADGER DON&apos;T CARE</h2>
        <p className={styles.ctaText}>About your runtime errors. Because there won&apos;t be any.</p>
        <Link className={styles.btnPrimary} to="/docs/getting-started/installation">
          Read the Docs
        </Link>
      </section>
    </Layout>
  );
}
