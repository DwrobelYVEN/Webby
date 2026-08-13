'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api';
import type { EventListing } from '@/types';

export default function EventsPage() {
  const [events, setEvents] = useState<EventListing[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    api
      .listEvents()
      .then((data) => setEvents(data as EventListing[]))
      .catch((err) => setError(err instanceof Error ? err.message : 'Failed to load events'))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div>
      <h1>Find an Opportunity</h1>

      {loading && <p>Loading events…</p>}
      {error && (
        <div className="card">
          <p>Couldn't load events: {error}</p>
          <p style={{ fontSize: 13, color: '#666' }}>
            (Expected until the backend is running — see backend/README or the root docker-compose.yml)
          </p>
        </div>
      )}

      {!loading && !error && events.length === 0 && (
        <p>No published events yet.</p>
      )}

      {events.map((event) => (
        <div className="card" key={event.id}>
          <h3>{event.title}</h3>
          <p>{event.description}</p>
          <p style={{ fontSize: 13, color: '#666' }}>
            {new Date(event.startsAt).toLocaleString()} · {event.remote ? 'Remote' : event.location}
            {' · '}
            {event.currentSignups}/{event.capacity} signed up
          </p>
        </div>
      ))}
    </div>
  );
}
