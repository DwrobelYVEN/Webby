// Mirrors backend/internal/models — kept hand-in-sync for now. If this
// drifts, consider generating these from the Go structs (e.g. via
// tygo or openapi-generator) as a follow-up.

export type ServiceLogState =
  | 'draft'
  | 'submitted'
  | 'verified'
  | 'rejected'
  | 'flagged'
  | 'archived';

export interface Volunteer {
  id: string;
  userId: string;
  school: string;
  gradeLevel: string;
  skills: string[];
  interests: string[];
  locationPreference: string;
  totalHoursVerified: number;
  totalHoursPending: number;
  eventsAttended: number;
  followingPaused: boolean;
}

export interface EventListing {
  id: string;
  organizationId: string;
  title: string;
  description: string;
  roleExpectations: string;
  requiredSkills: string[];
  startsAt: string;
  endsAt: string;
  remote: boolean;
  location: string;
  capacity: number;
  currentSignups: number;
}

export interface ServiceLog {
  id: string;
  entryId: string;
  eventId: string;
  rolePerformed: string;
  serviceDate: string;
  hoursServed: number;
  state: ServiceLogState;
}

export interface VSR {
  id: string;
  volunteerId: string;
  totalVerifiedHours: number;
  lastUpdatedAt: string;
  locked: boolean;
}
