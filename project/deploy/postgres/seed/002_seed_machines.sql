INSERT INTO machines (machine_name, brand, model, protocol, connection_type, ip_address, port, notes)
SELECT * FROM (VALUES
    ('Fluid Bed Dryer',             'Mitsubishi',       'FX5U-64M',       'mc',        'ethernet', '192.168.1.10', 5007, 'SLMP 3E Frame, Ethernet-based communication'),
    ('Fluid Bed Processor',         'Mitsubishi',       'FX5U-64M',       'mc',        'ethernet', '192.168.1.11', 5007, 'SLMP 3E Frame, Ethernet-based communication'),
    ('Fluid Bed Equipment',         'Mitsubishi',       'FX5U-64M',       'mc',        'ethernet', '192.168.1.12', 5007, 'SLMP 3E Frame, Ethernet-based communication'),
    ('Tablet Coating Machine',      'Mitsubishi',       'FX5U-64M',       'mc',        'ethernet', '192.168.1.13', 5007, 'SLMP 3E Frame, Ethernet-based communication'),
    ('Compression Machine',         'Omron',            'SYSMAC CJ2M CPU14', 'fins',   'ethernet', '192.168.1.20', 9600, 'FINS/TCP via ETN21 unit'),
    ('Compression Machine',         'B&R',              'X20 CP1686X',     'opcua',     'ethernet', '192.168.1.30', 4840, 'OPC UA server with global PVs'),
    ('Tablet Printing Machine',     'Allen Bradley',    'MicroLogix 1400', 'ethernetip','ethernet','192.168.1.40', 44818, 'CIP over EtherNet/IP'),
    ('Capsule Checkweigher',        'B&R',              'X20 BC0083',      'opcua',     'ethernet', '192.168.1.31', 4840, 'OPC UA bus controller'),
    ('Rapid Mixer Granulator',      'Mitsubishi',       'FX5U-64M',       'mc',        'ethernet', '192.168.1.14', 5007, 'SLMP 3E Frame, Ethernet-based communication'),
    ('Blender',                     'Mitsubishi',       'FX5U-32M',       'mc',        'ethernet', '192.168.1.15', 5007, 'SLMP 3E Frame, Ethernet-based communication'),
    ('Blender',                     'Mitsubishi',       'FX3GE-24M',      'mc',        'ethernet', '192.168.1.16', 5007, 'SLMP 3E Frame, built-in Ethernet on FX3GE')
) AS t(machine_name, brand, model, protocol, connection_type, ip_address, port, notes)
WHERE NOT EXISTS (SELECT 1 FROM machines);
