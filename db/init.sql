CREATE TABLE employee
(
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username   VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(50),
    last_name  VARCHAR(50),
    created_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

CREATE TYPE organization_type AS ENUM (
    'IE',
    'LLC',
    'JSC'
    );

CREATE TABLE organization
(
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name        VARCHAR(100) NOT NULL,
    description TEXT,
    type        organization_type,
    created_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE organization_responsible
(
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID REFERENCES organization (id) ON DELETE CASCADE,
    user_id         UUID REFERENCES employee (id) ON DELETE CASCADE
);

create type service_type as enum ('Construction', 'Delivery', 'Manufacture');
create type tender_status as enum ('Created', 'Published', 'Closed');

create table tender
(
    id              uuid      default uuid_generate_v4()                not null,
    name            varchar(100)                                        not null,
    description     varchar(500)                                        not null,
    service_type    service_type                                        not null,
    status          tender_status                                       not null,
    organization_id uuid REFERENCES organization (id) on delete cascade not null,
    created_at      timestamp default now()                             not null,
    updated_at      timestamp default now()                             not null,
    creator_id      uuid references employee (id)                       not null,
    version         integer   default 1                                 not null,
    primary key (id, version)
);