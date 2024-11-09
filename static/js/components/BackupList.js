const { 
    Table, Input, Card, message, Spin, Button, 
    Modal, Tag, Space, Typography, Tooltip,
    ExportOutlined  // 添加图标组件
} = window.antdComponents;

const { Search } = Input;
const { Text } = Typography;

function BackupList() {
    console.log('BackupList component rendering'); // 添加日志
    const [backups, setBackups] = React.useState([]);
    const [awsSnapshots, setAwsSnapshots] = React.useState([]);
    const [loading, setLoading] = React.useState(false);
    const [searchText, setSearchText] = React.useState('');
    const [exportModalVisible, setExportModalVisible] = React.useState(false);
    const [selectedSnapshot, setSelectedSnapshot] = React.useState(null);

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
        console.log('useEffect triggered'); // 添加日志
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

            console.log('Starting to fetch data'); // 添加日志
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
            title: '最近一次备份时间',
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
                <Tooltip title="点击下载备份文件">
                    <Button 
                        type="link" 
                        onClick={() => {
                            Modal.confirm({
                                title: '确认下载',
                                content: '您确定要下载此备份文件吗？文件可能较大，请确保网络环境良好。',
                                onOk: () => window.open(text, '_blank'),
                                okText: '确认下载',
                                cancelText: '取消'
                            });
                        }}
                    >
                        下载
                    </Button>
                </Tooltip>
            ) : (
                <Tag color="warning">无备份</Tag>
            ),
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
            title: '最近一次备份时间',
            dataIndex: 'snapshotCreateTime',
            key: 'snapshotCreateTime',
            render: (text) => text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-',
            sorter: (a, b) => dayjs(a.snapshotCreateTime).unix() - dayjs(b.snapshotCreateTime).unix(),
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status) => getStatusTag(status),
        },
        {
            title: '区域',
            dataIndex: 'region',
            key: 'region',
            render: (region) => (
                <Tag color="blue">{region}</Tag>
            ),
        },
        {
            title: '操作',
            key: 'action',
            render: (_, record) => (
                <Space>
                    <Tooltip title={
                        record.status !== 'available' 
                            ? '只有状态为可用的快照才能导出' 
                            : '导出快照到 S3'
                    }>
                        <Button 
                            type="primary"
                            className="action-button"
                            icon={<ExportOutlined />}
                            onClick={() => showExportConfirm(record)}
                            disabled={record.status !== 'available'}
                        >
                            导出快照
                        </Button>
                    </Tooltip>
                </Space>
            ),
        },
    ];

    const filteredBackups = backups.filter(backup => 
        backup.env.toLowerCase().includes(searchText.toLowerCase())
    );

    const filteredSnapshots = awsSnapshots.filter(snapshot => 
        snapshot.env.toLowerCase().includes(searchText.toLowerCase())
    );

    const showExportConfirm = (record) => {
        setSelectedSnapshot(record);
        setExportModalVisible(true);
    };

    const handleExportConfirm = async () => {
        try {
            setExportModalVisible(false);
            message.loading({
                content: '正在启动导出任务...',
                key: 'exportMessage',
                duration: 0
            });

            const response = await fetch(`/awsrds/export/${selectedSnapshot.env}`, {
                method: 'POST'
            });
            const data = await response.json();
            
            if (response.ok) {
                message.success({
                    content: (
                        <div>
                            <div>快照导出任务已启动</div>
                            <div>任务ID: {data.export_task_id}</div>
                            <div>目标位置: s3://{data.s3_bucket}/{data.s3_prefix}</div>
                        </div>
                    ),
                    key: 'exportMessage',
                    duration: 5
                });
            } else {
                message.error({
                    content: `启动导出任务失败: ${data.error}`,
                    key: 'exportMessage'
                });
            }
        } catch (error) {
            message.error({
                content: `请求失败: ${error.message}`,
                key: 'exportMessage'
            });
        }
    };

    const getStatusTag = (status) => {
        const statusConfig = {
            'available': { color: 'success', text: '可用' },
            'creating': { color: 'processing', text: '创建中' },
            'failed': { color: 'error', text: '失败' },
            'deleting': { color: 'warning', text: '删除中' }
        };
        const config = statusConfig[status] || { color: 'default', text: status };
        return (
            <Tag color={config.color} className="status-tag">
                {config.text}
            </Tag>
        );
    };

    return (
        <div>
            <Search
                placeholder="搜索环境名称"
                allowClear
                enterButton
                className="search-box"
                onChange={(e) => setSearchText(e.target.value)}
            />
            <Spin spinning={loading}>
                <div className="card-container">
                    <Card 
                        title={<Text strong>阿里云 RDS 备份</Text>}
                        className="custom-card"
                        style={{ marginBottom: 24 }}
                    >
                        <Table
                            columns={aliColumns}
                            dataSource={filteredBackups}
                            pagination={false}
                            rowKey="env"
                        />
                    </Card>
                    <Card 
                        title={<Text strong>AWS RDS 快照</Text>}
                        className="custom-card"
                    >
                        <Table
                            columns={awsColumns}
                            dataSource={filteredSnapshots}
                            pagination={false}
                            rowKey="env"
                        />
                    </Card>
                </div>
            </Spin>

            <Modal
                title="确认导出快照"
                visible={exportModalVisible}
                onOk={handleExportConfirm}
                onCancel={() => setExportModalVisible(false)}
                okText="确认导出"
                cancelText="取消"
            >
                {selectedSnapshot && (
                    <div>
                        <p>您确定要导出以下快照吗？</p>
                        <p><strong>环境：</strong>{selectedSnapshot.env}</p>
                        <p><strong>快照ID：</strong>{selectedSnapshot.snapshotId}</p>
                        <p><strong>创建时间：</strong>{dayjs(selectedSnapshot.snapshotCreateTime).format('YYYY-MM-DD HH:mm:ss')}</p>
                        <p><strong>区域：</strong>{selectedSnapshot.region}</p>
                    </div>
                )}
            </Modal>
        </div>
    );
}

window.BackupList = BackupList;
