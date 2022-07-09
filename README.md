# Schedule
This package is a dependency free utility package.  It is a collection of time related value objects.  The main reason it exists is because a `time.Time` is one specific moment in time.  It is a Date, Clock, and Timezone all in one.

Sometimes we just want a `Date` or `DateRange` which does not care about times or time zones.  Sometimes we just want a `Clock` or `TimeSlot` which does not care about dates.  We wanted to know if a `Date` was within a `DateRange` without having to worry about what timezone it is in and making sure the start date was 00:00 end date was at 23:59.  It is a date, the time doesn't matter in these cases.

## Clock
A time without date information which will always be within 24 hours.  If you add minutes to it that go past 24 hours it auto wraps as a clock would.

It json and sql encodes/decodes to/from a "HH:MM" string

Why not use `time.Time` ?  Because a `time.Time` is much more than we need.  We just need a clock, that does not care about timezones or dates.  If you store a timeslot with `time.Time` 's then those objects have dates and timezone information in them.

String format: "13:00"

### Constructors
```
  NewClock(h, m int) Clock
  ParseClock(string) Clock // from "HH:MM" format
```

### Useful Methods
```
  String() string             // "HH:MM" format
  Add(minutes int) Clock
  Subtract(mintues int) Clock
  Hour() int
  Minute() int
  Equal(Clock) bool
  Before(Clock) bool
  After(Clock) bool
  IsZero() bool
  Pointer() *Clock            // useful for setting Until on some other types
  ToDuration() time.Duration
  ToTime(Date, *time.Location) time.Time
```

## TimeSlot
A `TimeSlot` is just a type with two `Clock`'s in it `Start` and `End`

It is important to note that a `TimeSlot` of "23:00-02:00" is valid and simply means it crosses midnight.  So it would have a 3 hour duration.

Also two time slots that share a single moment in time at its edge do not overlap.  So "09:00-10:00" and "10:00-11:00" time slots do not conflict, they just share a boundary.

String format: "09:00-13:00"

### Constructors
```
  NewTimeSlot(start, end Clock) TimeSlot
  ParseTimeSlot(string) TimeSlot         // from "HH:MM-HH:MM" format
```

### Useful Methods
```
  String() string   // "HH:MM-HH:MM" format
  StartTime() Clock
  EndTime() Clock
  Minutes() int
  Duration() time.Duration
  IsZero() bool
```

## Weekday
More or less the same as `time.Weekday` other than it json/sql encode/decode 's to/from string name.  Which is the only reason this type exists really.

```
type Weekday time.Weekday

const (
	Sunday    = Weekday(time.Sunday)
	Monday    = Weekday(time.Monday)
	Tuesday   = Weekday(time.Tuesday)
	Wednesday = Weekday(time.Wednesday)
	Thursday  = Weekday(time.Thursday)
	Friday    = Weekday(time.Friday)
	Saturday  = Weekday(time.Saturday)
)
```

### Constructors
```
  ParseWeekday(string) (Weekday, error)
```

### Useful Methods
```
  String() string
  Next() Weekday
```

## WeekdayTimeSlot
String format: "Monday 09:00-13:00"

A `WeekdayTimeSlot` can be converted to/from an int.  The purpose of this is simply to provide something like a unique key to this value object.  The int is actually the weekday, start, and end all encoded into an int.  However nothing that uses the int value should understand how it is encoded or care, it is just used sometimes as a unique key to identify a value.  It is also helpful for sorting and knowing if two values are equal or not.

### Constructors
```
  NewWeekdayTimeSlot(Weekday, TimeSlot) WeekdayTimeSlot
  NewWeekdayAllDayTimeSlot(Weekday) WeekdayTimeSlot
  WeekdayTimeSlotFromString(string) WeekdayTimeSlot
  WeekdayTimeSlotFromInt(int) WeekdayTimeSlot
```

### Methods
```
  Weekday() Weekday
  Slot() TimeSlot
  String() string
  ToInt() int
  Start() Clock
  End() Clock
  Minutes() int
  Duration() time.Duration
  IsAllDay() bool
  OverlapsWith(WeekdayTimeSlot) bool
  Equal() bool
```

### Helper functions
```
  SortWeekdayTimeSlots(...WeekdayTimeSlot) []WeekdayTimeSlot
  UniqueWeekdayTimeSlots(...WeekdayTimeSLot) []WeekdayTimeSlot
  SlotKeys(...WeekdayTimeSlot) []int
```

## WeekdayTimeSlotMap
This is really the same thing as a `[]WeekdayTimeSlot` but organized as `map[Weekday][]TimeSlot`

It can be easier to work with at times and has unique enforcement built in.  It has methods to convert this to/from `[]WeekdayTimeSlot` so you can go back and forth as needed.  Though be warned that you will lose any duplicates, though I've yet to actually want any duplicates.

### Constructors
```
  NewWeekdayTimeSlotMap() WeekdayTimeSlotMap
  WeekdayTimeSlotMapFromSlice([]WeekdayTimeSlot) WeekdayTimeSlotMap
```

### Methods
```
  Add(Weekday, ...TimeSlot) WeekdayTimeSlotMap
  Has(Weekday, ...TimeSlot) bool
  AddTimeSlot(day Weekday, start, end Clock) WeekdayTimeSlotMap
  TimeSlots(day Weekday) []TimeSlot
  ToWeekdayTimeSlots() []WeekdayTimeSlot
```

## Date
A date is any calendar date.  It takes advantage of `time.Time` to do anything complicated but it has no time or location data built into it.

json/sql encode/decode to/from string

String format: "2022-07-09"

### Constructors
```
  Today() Date
  NewDate(year int, month time.Month, day int) Date
  NewDateFromTime(t time.Time) Date
  ParseDate(string) *Date
  ZeroDate() *Date
```

### Methods
```
  String() string
  Year() int
  Month() time.Month
  Day() int
  Weekday() Weekday
  Before(Date) bool
  After(Date) bool
  Equal(Date) bool
  Next() Date
  Pointer() *Date
  IsZero() bool
  AddDate(year, month, day int) Date
  Sub(Date) int
  ToTime() time.Time
```

## DateRange
A `DateRange` goes from `Date` until `*Date` and so the until can be `nil` and when it is `nil` it means that the `DateRange` has no end and therefore is interpreted as "forever".  It is important to understand this when reasoning how Overlap or Contains work.

Also a `DateRange` is inclusive at both ends, check out this comment from the code:

```
// DateRange represents a set of days
// this set includes the From date and the Until date
// from Jan1 until Jan1 is one day  
// from Jan1 until Jan2 is two days  
// when Until is nil it means forever
```
Isn't it nice to not have to worry about time when you don't need to?

```
// InfDays represents an infinite number of days
// MaxInt32 days is over 5.8 million years  
// we are not using a negative number because someone  
// may check DayCount() > 0 to know if it has days or not  
const InfDays = math.MaxInt32
```

### Constructors
```
  NewDateRange() DateRange   // Today until forever
  NewDateRangeUntil(from Date, until *Date) DateRange
  ZeroDateRange() DateRange
```

### Methods
```
  WithFrom(Date) DateRange
  WithUntil(Date) DateRange
  Validate() error          // from is required, from can not be after until
  IsZero() bool
  ContainsDate(Date)
  Overlaps(DateRange) bool
  Exceeds(DateRange) bool
  Equal(DateRange) bool
  String() string           // "from 2022-01-01 until forever"
  HasDays() bool            // DayCount() > 0
  DayCount() int            // when until is nil then returns InfDays
```

## Schedule
```
type Schedule struct {
	DateRange DateRange
	TimeSlots []WeekdayTimeSlot
}
```

### Constructors
```
  NewSchedule(dr DateRange, slots ...WeekdayTimeSlot)
```

### Methods
```
  WithDateRange(dateRange DateRange) Schedule
  WithFrom(from Date) Schedule
  WithUntil(until Date) Schedule
  WithTimeSlots(slots ...WeekdayTimeSlot) Schedule
  From() Date
  Until() *Date
  IsEmpty() bool
  HasTimeSlots() bool
  Merge(schedules ...Schedule) Schedule
```

#### Merge

Before using this function you should really understand what it does...

```
// Merge does a merge on both the schedule dateRanges and the timeslots
//   The intended use for this is to merge a parent schedule with a sub schedule
//   the sub schedule is a subset of the parent, however it may have all-day entries
//   which allow for config on "any" timeslots for that day
//   if timeslots exist on only the parent, they are excluded
//   if timeslots exist on only the child, they are invalid, and therefore excluded
// dateRange
//   a merged dateRange is where they intersect
//   - ex: jan01-jan30 merged with jan15-feb15 results in jan15-jan30
// timeslots
//   are only kept in a merge when they exist in all schedules without conflicting
//   exception is when one schedule has an all day timeslot and another specific timeslots
//   this results in the specific timeslots being kept in favor of the all day
//   - ex: Mon7-8,Tues8-9,Wed merged with Mon7-8,Tues6-7,Wed7-8 results in Mon7-8,Wed7-8
//         Tues809,Tues6-7 were excluded because they only exist in one schedule
//         Wed7-8 was included because the other schedule was for Wed all day
// see TestSchedulesMerge for a good example
```

## Calendar

### Constructors
```
  NewCalendar(schedules ...Schedule) Calendar
```

### Methods
```
  WithSchedules(schedules ...Schedule) Calendar
  ByDate(limit Date) CalendarMap
```

## CalendarMap
```
type CalendarMap map[Date][]WeekdayTimeSlot
```

### Methods
```
HasDate(date Date) bool
```