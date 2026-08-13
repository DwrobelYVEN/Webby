import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'YVEN — Youth Volunteer Engagement Network',
  description: 'Volunteer service logging, verification, and VSR issuance.',
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <header className="site-nav">
          <a href="/" className="brand">YVEN</a>
          <nav>
            <a href="/events">Find Events</a>
            <a href="/volunteer/signup">Volunteer Sign Up</a>
            <a href="/dashboard">Dashboard</a>
          </nav>
        </header>
        <main>{children}</main>
      </body>
    </html>
  );
}
