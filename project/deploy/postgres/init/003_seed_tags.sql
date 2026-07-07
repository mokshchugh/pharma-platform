INSERT INTO tags (machine_id, tag_name, description, data_type, scale_factor, unit, address)
SELECT * FROM (VALUES
    -- Machine 1: Fluid Bed Dryer (Mitsubishi FX5U-64M)
    (1, 'Inlet_Air_Temp',          'Inlet Air Temperature',          'int16',  0.1,  '°C',    'D100'),
    (1, 'Outlet_Air_Temp',         'Outlet Air Temperature',         'int16',  0.1,  '°C',    'D102'),
    (1, 'Product_Temp',            'Product Temperature',            'int16',  0.1,  '°C',    'D104'),
    (1, 'Differential_Pressure',   'Differential Pressure',          'int16',  0.1,  'mbar',  'D106'),
    (1, 'Fan_Speed',               'Fan Speed',                      'int16',  1.0,  'Hz',    'D108'),
    (1, 'Damper_Position',         'Damper Position',                'int16',  1.0,  '%',     'D110'),
    (1, 'Bag_Shake_Count',         'Bag Shake Counter',              'int16',  1.0,  'count', 'D112'),
    (1, 'Batch_ID',                'Batch Identifier',               'int16',  1.0,  '',      'D114'),
    (1, 'Process_Timer_Elapsed',   'Process Timer (Elapsed)',        'int32',  1.0,  'sec',   'D116'),
    (1, 'Run_Status',              'Run Status',                     'bool',   1.0,  '',      'M100'),
    (1, 'Alarm_Status',            'Alarm Status',                   'bool',   1.0,  '',      'M101'),
    (1, 'Door_Interlock_Status',   'Door Interlock Status',          'bool',   1.0,  '',      'M102'),

    -- Machine 2: Fluid Bed Processor (Mitsubishi FX5U-64M)
    (2, 'Inlet_Air_Temp',          'Inlet Air Temperature',          'int16',  0.1,  '°C',    'D100'),
    (2, 'Outlet_Air_Temp',         'Outlet Air Temperature',         'int16',  0.1,  '°C',    'D102'),
    (2, 'Product_Temp',            'Product Temperature',            'int16',  0.1,  '°C',    'D104'),
    (2, 'Spray_Rate',              'Spray Rate',                     'int16',  0.1,  'ml/min','D106'),
    (2, 'Atomization_Air_Pressure','Atomization Air Pressure',       'int16',  0.01, 'bar',   'D108'),
    (2, 'Peristaltic_Pump_Speed',  'Peristaltic Pump Speed',         'int16',  1.0,  'rpm',   'D110'),
    (2, 'Differential_Pressure',   'Differential Pressure',          'int16',  0.1,  'mbar',  'D112'),
    (2, 'Batch_ID',                'Batch Identifier',               'int16',  1.0,  '',      'D114'),
    (2, 'Process_Phase',           'Process Phase',                  'int16',  1.0,  '',      'D116'),
    (2, 'Run_Status',              'Run Status',                     'bool',   1.0,  '',      'M100'),
    (2, 'Alarm_Status',            'Alarm Status',                   'bool',   1.0,  '',      'M101'),

    -- Machine 3: Fluid Bed Equipment (Mitsubishi FX5U-64M)
    (3, 'Inlet_Air_Temp',          'Inlet Air Temperature',          'int16',  0.1,  '°C',    'D100'),
    (3, 'Outlet_Air_Temp',         'Outlet Air Temperature',         'int16',  0.1,  '°C',    'D102'),
    (3, 'Blower_Speed',            'Blower Speed',                   'int16',  1.0,  'rpm',   'D104'),
    (3, 'Differential_Pressure',   'Differential Pressure',          'int16',  0.1,  'mbar',  'D106'),
    (3, 'Filter_DP_Alarm_Setpoint','Filter DP Alarm Setpoint',       'int16',  0.1,  'mbar',  'D108'),
    (3, 'Run_Status',              'Run Status',                     'bool',   1.0,  '',      'M100'),
    (3, 'Alarm_Status',            'Alarm Status',                   'bool',   1.0,  '',      'M101'),
    (3, 'Maintenance_Due_Flag',    'Maintenance Due Flag',           'bool',   1.0,  '',      'M102'),

    -- Machine 4: Tablet Coating Machine (Mitsubishi FX5U-64M)
    (4, 'Pan_Speed',               'Pan Speed',                      'int16',  1.0,  'rpm',   'D100'),
    (4, 'Spray_Rate',              'Spray Rate',                     'int16',  0.1,  'ml/min','D102'),
    (4, 'Inlet_Air_Temp',          'Inlet Air Temperature',          'int16',  0.1,  '°C',    'D104'),
    (4, 'Exhaust_Air_Temp',        'Exhaust Air Temperature',        'int16',  0.1,  '°C',    'D106'),
    (4, 'Product_Temp',            'Product Temperature',            'int16',  0.1,  '°C',    'D108'),
    (4, 'Gun_Air_Pressure',        'Gun Air Pressure',               'int16',  0.01, 'bar',   'D110'),
    (4, 'Peristaltic_Pump_Speed',  'Peristaltic Pump Speed',         'int16',  1.0,  'rpm',   'D112'),
    (4, 'Weight_Gain_Percent',     'Weight Gain Percent',            'int16',  0.1,  '%',     'D114'),
    (4, 'Process_Timer_Elapsed',   'Process Timer Elapsed',          'int32',  1.0,  'sec',   'D116'),
    (4, 'Run_Status',              'Run Status',                     'bool',   1.0,  '',      'M100'),
    (4, 'Alarm_Status',            'Alarm Status',                   'bool',   1.0,  '',      'M101'),

    -- Machine 5: Compression Machine (Omron CJ2M)
    (5, 'Main_Compression_Force',  'Main Compression Force',         'int16',  1.0,  'kN',    'D100'),
    (5, 'PreCompression_Force',    'Pre-Compression Force',          'int16',  1.0,  'kN',    'D102'),
    (5, 'Machine_Speed',           'Machine Speed',                  'int16',  1.0,  'tpm',   'D104'),
    (5, 'Ejection_Force',          'Ejection Force',                 'int16',  1.0,  'N',     'D106'),
    (5, 'Upper_Punch_Penetration', 'Upper Punch Penetration',        'int16',  1.0,  'mm',    'D108'),
    (5, 'Avg_Tablet_Weight',       'Avg Tablet Weight',              'int16',  1.0,  'mg',    'D110'),
    (5, 'Good_Count',              'Good Tablet Count',              'int32',  1.0,  'count', 'D112,D113'),
    (5, 'Reject_Count',            'Reject Tablet Count',            'int32',  1.0,  'count', 'D114,D115'),
    (5, 'Hopper_Level',            'Hopper Level',                   'int16',  1.0,  '%',     'D116'),
    (5, 'Run_Status',              'Run Status',                     'bool',   1.0,  '',      'CIO100.00'),
    (5, 'Alarm_Status',            'Alarm Status',                   'bool',   1.0,  '',      'CIO100.01'),

    -- Machine 6: Compression Machine (B&R X20 CP1686X)
    (6, 'MainCompForce',           'Main Compression Force',         'float32',1.0,  'kN',    'ns=6;s="Application"."gPV"."MainCompForce"'),
    (6, 'PreCompForce',            'Pre-Compression Force',          'float32',1.0,  'kN',    'ns=6;s="Application"."gPV"."PreCompForce"'),
    (6, 'TurretSpeed',             'Turret Speed',                   'float32',1.0,  'tpm',   'ns=6;s="Application"."gPV"."TurretSpeed"'),
    (6, 'EjectionForce',           'Ejection Force',                 'float32',1.0,  'N',     'ns=6;s="Application"."gPV"."EjectionForce"'),
    (6, 'GoodCount',               'Good Tablet Count',              'int32',  1.0,  'count', 'ns=6;s="Application"."gPV"."GoodCount"'),
    (6, 'RejectCount',             'Reject Tablet Count',            'int32',  1.0,  'count', 'ns=6;s="Application"."gPV"."RejectCount"'),
    (6, 'HopperLevel',             'Hopper Level',                   'float32',1.0,  '%',     'ns=6;s="Application"."gPV"."HopperLevel"'),
    (6, 'RunStatus',               'Run Status',                     'bool',   1.0,  '',      'ns=6;s="Application"."gPV"."RunStatus"'),
    (6, 'AlarmStatus',             'Alarm Status',                   'bool',   1.0,  '',      'ns=6;s="Application"."gPV"."AlarmStatus"'),

    -- Machine 7: Tablet Printing Machine (Allen Bradley MicroLogix 1400)
    (7, 'Machine_Speed',           'Machine Speed',                  'int16',  1.0,  'tpm',   'N7:0'),
    (7, 'Good_Print_Count',        'Good Print Count',               'int16',  1.0,  'count', 'N7:1'),
    (7, 'Reject_Count',            'Reject Count',                   'int16',  1.0,  'count', 'N7:2'),
    (7, 'Ink_Level',               'Ink Level',                      'int16',  1.0,  '%',     'N7:3'),
    (7, 'Batch_ID',                'Batch Identifier',               'int16',  1.0,  '',      'N7:4'),
    (7, 'Run_Status',              'Run Status',                     'bool',   1.0,  '',      'B3:0/0'),
    (7, 'Alarm_Status',            'Alarm Status',                   'bool',   1.0,  '',      'B3:0/1'),
    (7, 'PrintHead_Fault',         'Print Head Fault',               'bool',   1.0,  '',      'B3:0/2'),

    -- Machine 8: Capsule Checkweigher (B&R X20 BC0083)
    (8, 'CheckWeight',             'Check Weight',                   'float32',1.0,  'mg',    'ns=6;s="Application"."gPV"."CheckWeight"'),
    (8, 'Good_Count',              'Good Count',                     'int32',  1.0,  'count', 'ns=6;s="Application"."gPV"."GoodCountCw"'),
    (8, 'Reject_Count',            'Reject Count',                   'int32',  1.0,  'count', 'ns=6;s="Application"."gPV"."RejectCountCw"'),
    (8, 'Line_Speed',              'Line Speed',                     'float32',1.0,  'm/s',   'ns=6;s="Application"."gPV"."LineSpeed"'),
    (8, 'Reject_Reason_Code',      'Reject Reason Code',             'int16',  1.0,  '',      'ns=6;s="Application"."gPV"."RejectReasonCode"'),
    (8, 'Run_Status',              'Run Status',                     'bool',   1.0,  '',      'ns=6;s="Application"."gPV"."RunStatusCw"'),
    (8, 'Alarm_Status',            'Alarm Status',                   'bool',   1.0,  '',      'ns=6;s="Application"."gPV"."AlarmStatusCw"'),

    -- Machine 9: Rapid Mixer Granulator (Mitsubishi FX5U-64M)
    (9, 'Impeller_Speed',           'Impeller Speed',                 'int16',  1.0,  'rpm',   'D100'),
    (9, 'Chopper_Speed',            'Chopper Speed',                  'int16',  1.0,  'rpm',   'D102'),
    (9, 'Binder_Addition_Rate',     'Binder Addition Rate',           'int16',  0.1,  'ml/min','D104'),
    (9, 'Product_Temp',             'Product Temperature',            'int16',  0.1,  '°C',    'D106'),
    (9, 'Impeller_Motor_Load',      'Impeller Motor Load',            'int16',  1.0,  '%',     'D108'),
    (9, 'Kneading_Timer_Elapsed',   'Kneading Timer Elapsed',         'int16',  1.0,  'sec',   'D110'),
    (9, 'Batch_ID',                 'Batch Identifier',               'int16',  1.0,  '',      'D112'),
    (9, 'Process_Phase',            'Process Phase',                  'int16',  1.0,  '',      'D114'),
    (9, 'Run_Status',               'Run Status',                     'bool',   1.0,  '',      'M100'),
    (9, 'Alarm_Status',             'Alarm Status',                   'bool',   1.0,  '',      'M101'),

    -- Machine 10: Blender (Mitsubishi FX5U-32M)
    (10, 'Blender_RPM',             'Blender RPM',                    'int16',  1.0,  'rpm',   'D100'),
    (10, 'Blend_Timer_Elapsed',     'Blend Timer Elapsed',            'int16',  1.0,  'sec',   'D102'),
    (10, 'Batch_ID',                'Batch Identifier',               'int16',  1.0,  '',      'D104'),
    (10, 'Load_Cell_Weight',        'Load Cell Weight',               'int16',  1.0,  'kg',    'D106'),
    (10, 'Door_Interlock_Status',   'Door Interlock Status',          'bool',   1.0,  '',      'M100'),
    (10, 'Run_Status',              'Run Status',                     'bool',   1.0,  '',      'M101'),
    (10, 'Alarm_Status',            'Alarm Status',                   'bool',   1.0,  '',      'M102'),

    -- Machine 11: Blender (Mitsubishi FX3GE-24M)
    (11, 'Blender_RPM',             'Blender RPM',                    'int16',  1.0,  'rpm',   'D100'),
    (11, 'Blend_Timer_Elapsed',     'Blend Timer Elapsed',            'int16',  1.0,  'sec',   'D102'),
    (11, 'Batch_ID',                'Batch Identifier',               'int16',  1.0,  '',      'D104'),
    (11, 'Load_Cell_Weight',        'Load Cell Weight',               'int16',  1.0,  'kg',    'D106'),
    (11, 'Run_Status',              'Run Status',                     'bool',   1.0,  '',      'M100'),
    (11, 'Alarm_Status',            'Alarm Status',                   'bool',   1.0,  '',      'M101')
) AS t(machine_id, tag_name, description, data_type, scale_factor, unit, address)
WHERE NOT EXISTS (SELECT 1 FROM tags);
