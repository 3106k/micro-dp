DELETE FROM plans
WHERE id IN (
    'plan-free-default',
    'plan-starter-default',
    'plan-pro-default'
);
