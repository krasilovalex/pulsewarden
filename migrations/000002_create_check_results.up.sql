CREATE TABLE check_results (
    id UUID PRIMARY KEY,


    monitor_id UUID NOT NULL
        REFERENCES monitors(id)
        ON DELETE CASCADE,

    status TEXT NOT NULL,
    status_code SMALLINT,
    latency_milliseconds INTEGER,
    error_message TEXT,
    checked_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT check_results_status_supported
        CHECK (status IN ('up', 'down')),

    CONSTRAINT check_results_status_code_valid
        CHECK (
            status_code IS NULL
            OR (
                status_code >= 100
                AND status_code <= 599
            )
        ),
    
    CONSTRAINT check_results_latency_non_negative
        CHECK(latency_milliseconds >= 0),

    CONSTRAINT check_results_up_valid
        CHECK (
            status <> 'up'
            OR (
                status_code IS NOT NULL
                AND error_message IS NULL
            )
        ),
    
    CONSTRAINT check_results_down_valid
        CHECK (
            status <> 'down'
            OR (
                status_code IS NOT NULL
                OR error_message IS NOT NULL
            )
        )
);

CREATE INDEX check_results_monitor_checked_at_idx
    ON check_results (
        monitor_id,
        checked_at DESC,
        id DESC
    );