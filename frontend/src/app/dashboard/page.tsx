'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api';
import type { Volunteer } from '@/types';

export default function DashboardPage() {
  const [volunteer, setVolunteer] = useState<Volunteer | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api
      .myDashboard()
      .then((data) => setVolunteer(data as Volunteer))
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load dashboard'))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <p>Loading…</p>;

  if (error) {
    return (
      <div className="card">
        <p>Couldn't load your dashboard: {error}</p>
        <p style={{ fontSize: 13, color: '#666' }}>
          (Expected until Auth0 session wiring is complete — see frontend/src/lib/api.ts)
        </p>
      </div>
    );
  }

  return (
    <div>
      <h1>Your Dashboard</h1>
      <div className="card">
        <h3>Progress</h3>
        <p>Verified hours: <strong>{volunteer?.totalHoursVerified ?? 0}</strong></p>
        <p>Pending hours: <strong>{volunteer?.totalHoursPending ?? 0}</strong></p>
        <p>Events attended: <strong>{volunteer?.eventsAttended ?? 0}</strong></p>
      </div>
      <div className="card">
        <h3>Quick links</h3>
        <p><a href="/events">Find new events</a></p>
        <p><a href="/dashboard/vsr">View / export your VSR</a></p>
      </div>
    </div>
  );
}
