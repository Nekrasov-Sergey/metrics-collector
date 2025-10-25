do
$$
    begin
        create type metric_type as enum ('gauge', 'counter');
    exception
        when duplicate_object then null;
    end
$$;

create table if not exists metrics
(
    name  text        not null primary key,
    type  metric_type not null,
    delta bigint,
    value double precision
);
