'use client';

import { useState } from 'react';
import { api } from '@/lib/api';

export default function VolunteerSignupPage() {
  const [status, setStatus] = useState<'idle' | 'submitting' | 'success' | 'error'>('idle');
  const [errorMsg, setErrorMsg] = useState('');

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setStatus('submitting');
    const form = new FormData(e.currentTarget);

    // NOTE: auth0Sub below is a placeholder. In the real flow, this page
    // sits behind the Auth0 signup redirect — the sub claim comes from
    // that session, not a hidden form field.
    const payload = {
      fullName: form.get('fullName'),
      email: form.get('email'),
      phone: form.get('phone'),
      school: form.get('school'),
      gradeLevel: form.get('gradeLevel'),
      emergencyContact: form.get('emergencyContact'),
      auth0Sub: 'TODO-from-auth0-session',
    };

    try {
      await api.registerVolunteer(payload);
      setStatus('success');
    } catch (err) {
      setStatus('error');
      setErrorMsg(err instanceof Error ? err.message : 'Something went wrong');
    }
  }

  if (status === 'success') {
    return <div className="card">You're registered. Check your email to finish setting up your account.</div>;
  }

  return (
    <div>
      <h1>Volunteer Sign Up</h1>
      <form onSubmit={handleSubmit}>
        <div className="field">
          <label htmlFor="fullName">Full legal name</label>
          <input id="fullName" name="fullName" required />
        </div>
        <div className="field">
          <label htmlFor="email">Email</label>
          <input id="email" name="email" type="email" required />
        </div>
        <div className="field">
          <label htmlFor="phone">Mobile phone</label>
          <input id="phone" name="phone" type="tel" />
        </div>
        <div className="field">
          <label htmlFor="school">School</label>
          <input id="school" name="school" />
        </div>
        <div className="field">
          <label htmlFor="gradeLevel">Grade level</label>
          <input id="gradeLevel" name="gradeLevel" />
        </div>
        <div className="field">
          <label htmlFor="emergencyContact">Emergency contact</label>
          <input id="emergencyContact" name="emergencyContact" required />
        </div>

        {status === 'error' && <p style={{ color: 'crimson' }}>{errorMsg}</p>}

        <button className="btn" type="submit" disabled={status === 'submitting'}>
          {status === 'submitting' ? 'Submitting…' : 'Create account'}
        </button>
      </form>
    </div>
  );
}
