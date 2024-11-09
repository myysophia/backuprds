const { Table, Input, Search, Card, message, Spin, Button } = window.antdComponents;

function BackupList() {
    console.log('BackupList component rendering');
    const [backups, setBackups] = React.useState([]);
    const [awsSnapshots, setAwsSnapshots] = React.useState([]);
    const [loading, setLoading] = React.useState(false);
    const [searchText, setSearchText] = React.useState('');

    const fetchBackup = async (env) => {
        console.log(`Fetching backup for ${env}`);
        try {
            const response = await fetch(`/alirds/${env}`);
            console.log(`Response for ${env}:`, response);
            const data = await response.json();
            console.log(`Data for ${env}:`, data);
            
            if (response.ok) {
                if (data.backup_download_url === "" && data.backup_intranet_download_url === "") {
                    message.warning(`${env} 环境无备份 (重试${data.retries}次)`);
                }
                setBackups(prev => [...prev, {
                    key: env,
                    env: env,
                    backupStartTime: data.backup_start_time || '-',
                    backupDownloadUrl: data.backup_download_url,
                    backupIntranetDownloadUrl: data.backup_intranet_download_url
                }]);
            } else {
                message.error(`获取 ${env} 环境备份失败: ${data.error} (重试${data.retries}次)`);
            }
        } catch (error) {
            console.error(`Error fetching backup for ${env}:`, error);
            message.error(`请求失败: ${error.message}`);
        }
    };

    const fetchAwsSnapshot = async (env) => {
        try {
            const response = await fetch(`/awsrds/${env}`);
            const data = await response.json();
            
            if (response.ok) {
                setAwsSnapshots(prev => [...prev, {
                    key: env,
                    env: env,
                    snapshotCreateTime: data.snapshot_create_time,
                    snapshotArn: data.snapshot_arn,
                    snapshotId: data.snapshot_id,
                    status: data.status,
                    instanceId: data.instance_id,
                    region: data.region
                }]);
            } else if (response.status === 404) {
                message.warning(`${env} 环境无快照`);
            } else {
                message.error(`获取 ${env} AWS快照失败: ${data.error}`);
            }
        } catch (error) {
            message.error(`请求失败: ${error.message}`);
        }
    };

    const exportAwsSnapshot = async (env) => {
        try {
            const response = await fetch(`/awsrds/export/${env}`, {
                method: 'POST'
            });
            const data = await response.json();
            
            if (response.ok) {
                message.success(`快照导出任务已启动: ${data.export_task_id}`);
            } else {
                message.error(`启动导出任务失败: ${data.error}`);
            }
        } catch (error) {
            message.error(`请求失败: ${error.message}`);
        }
    };

    React.useEffect(() => {
        console.log('useEffect triggered');
        const fetchData = async () => {
            setLoading(true);
            const aliEnvironments = [
                'vnnox-uat', 'vnnox-cn-db', 'vnnox-sg-db',
                'care-eu-db', 'vnnox-eu-db', 'care-us-db',
                'vnnox-us-db'
            ];
            const awsEnvironments = [
                'au-mysql8-care', 'au-mysql8-vnnox',
                'in-care-mysql', 'in-vnnox-mysql'
            ];

            console.log('Starting to fetch data');
            try {
                await Promise.all([
                    ...aliEnvironments.map(env => fetchBackup(env)),
                    ...awsEnvironments.map(env => fetchAwsSnapshot(env))
                ]);
            } catch (error) {
                console.error('Error fetching data:', error);
            } finally {
                setLoading(false);
            }
        };

        fetchData();
    }, []);

    const aliColumns = [
        {
            title: '环境名称',
            dataIndex: 'env',
            key: 'env',
            sorter: (a, b) => a.env.localeCompare(b.env),
        },
        {
            title: '备份开始时间',
            dataIndex: 'backupStartTime',
            key: 'backupStartTime',
            render: (text) => text === '-' ? '-' : dayjs(text).format('YYYY-MM-DD HH:mm:ss'),
            sorter: (a, b) => {
                if (a.backupStartTime === '-') return -1;
                if (b.backupStartTime === '-') return 1;
                return dayjs(a.backupStartTime).unix() - dayjs(b.backupStartTime).unix();
            },
        },
        {
            title: '公网下载链接',
            dataIndex: 'backupDownloadUrl',
            key: 'backupDownloadUrl',
            render: (text) => text ? (
                <a href={text} target="_blank" rel="noopener noreferrer">
                    下载
                </a>
            ) : '无备份',
        },
        {
            title: '内网下载链接',
            dataIndex: 'backupIntranetDownloadUrl',
            key: 'backupIntranetDownloadUrl',
            render: (text) => text ? (
                <a href={text} target="_blank" rel="noopener noreferrer">
                    下载
                </a>
            ) : '无备份',
        },
    ];

    const awsColumns = [
        {
            title: '环境名称',
            dataIndex: 'env',
            key: 'env',
            sorter: (a, b) => a.env.localeCompare(b.env),
        },
        {
            title: '快照创建时间',
            dataIndex: 'snapshotCreateTime',
            key: 'snapshotCreateTime',
            render: (text) => text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-',
            sorter: (a, b) => dayjs(a.snapshotCreateTime).unix() - dayjs(b.snapshotCreateTime).unix(),
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
        },
        {
            title: '区域',
            dataIndex: 'region',
            key: 'region',
        },
        {
            title: '操作',
            key: 'action',
            render: (_, record) => (
                <Button 
                    type="primary" 
                    onClick={() => exportAwsSnapshot(record.env)}
                    disabled={record.status !== 'available'}
                >
                    导出快照
                </Button>
            ),
        },
    ];

    const filteredBackups = backups.filter(backup => 
        backup.env.toLowerCase().includes(searchText.toLowerCase())
    );

    const filteredSnapshots = awsSnapshots.filter(snapshot => 
        snapshot.env.toLowerCase().includes(searchText.toLowerCase())
    );

    return (
        <div>
            <Search
                placeholder="搜索环境名称"
                allowClear
                enterButton
                style={{ marginBottom: 20 }}
                onChange={(e) => setSearchText(e.target.value)}
            />
            <Spin spinning={loading}>
                <Card title="阿里云 RDS 备份" style={{ marginBottom: 20 }}>
                    <Table
                        columns={aliColumns}
                        dataSource={filteredBackups}
                        pagination={false}
                        rowKey="env"
                    />
                </Card>
                <Card title="AWS RDS 快照">
                    <Table
                        columns={awsColumns}
                        dataSource={filteredSnapshots}
                        pagination={false}
                        rowKey="env"
                    />
                </Card>
            </Spin>
        </div>
    );
}

window.BackupList = BackupList;
