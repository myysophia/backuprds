console.log('BackupList component loading...');

// 错误边界组件
class ErrorBoundary extends React.Component {
    constructor(props) {
        super(props);
        this.state = { hasError: false, error: null };
    }

    static getDerivedStateFromError(error) {
        return { hasError: true, error };
    }

    componentDidCatch(error, errorInfo) {
        console.error('BackupList error:', error, errorInfo);
    }

    render() {
        if (this.state.hasError) {
            return <div>组件加载错误: {this.state.error.message}</div>;
        }
        return this.props.children;
    }
}

const {
    Table,
    message,
    Spin,
    Button,
    Modal,
    Tag,
    Space,
    Tooltip,
    Tabs,
    Card,
    Alert
} = window.antComponents;

const {
    ExportOutlined,
    CloudUploadOutlined,
    ReloadOutlined
} = window.antIcons;

function BackupList() {
    console.log('Initializing BackupList component...');
    
    // 状态管理
    const [loading, setLoading] = React.useState(false);
    const [activeTab, setActiveTab] = React.useState('1');
    const [aliData, setAliData] = React.useState([]);
    const [awsData, setAwsData] = React.useState([]);
    const [s3Config, setS3Config] = React.useState(null);
    const [instances, setInstances] = React.useState({
        aliyun: [],
        aws: []
    });

    // 添加轮询状态
    const [polling, setPolling] = React.useState(false);

    // 获取配置信息
    const fetchInstances = async () => {
        try {
            const response = await fetch('/instances');
            const data = await response.json();
            if (!data.error) {
                setInstances({
                    aliyun: data.aliyun || [],
                    aws: data.aws || []
                });
            } else {
                message.error('获取实例配置失败：' + data.error);
            }
        } catch (error) {
            console.error('获取实例配置失败：', error);
            message.error('获取实例配置失败');
        }
    };

    // 获取S3配置
    const fetchS3Config = async () => {
        try {
            const response = await fetch('/alirds/s3config');
            const data = await response.json();
            if (!data.error) {
                setS3Config(data);
            } else {
                console.error('获取S3配置失败：', data.error);
                message.error('获取S3配置失败：' + data.error);
            }
        } catch (error) {
            console.error('获取S3配置出错：', error);
            message.error('获取S3配置失败');
        }
    };

    // 阿里云数据获取
    const fetchAliData = async () => {
        try {
            setLoading(true);
            // 并行获取所有阿里云实例的数据
            const promises = instances.aliyun.map(async (instance) => {
                try {
                    const response = await fetch(`/alirds/${instance}`);
                    const data = await response.json();
                    
                    if (!data.error) {
                        return {
                            key: instance,
                            env: instance,
                            backup_start_time: data.backup_start_time,
                            backup_download_url: data.backup_download_url,
                            backup_intranet_download_url: data.backup_intranet_download_url,
                            retries: data.retries
                        };
                    } else {
                        console.error(`获取实例 ${instance} 数据失败:`, data.error);
                        return null;
                    }
                } catch (error) {
                    console.error(`获取实例 ${instance} 数据出错:`, error);
                    return null;
                }
            });

            const results = await Promise.all(promises);
            setAliData(results.filter(item => item !== null));
        } catch (error) {
            console.error('获取阿里云备份列表出错：', error);
            message.error('获取阿里云备份列表失败');
        } finally {
            setLoading(false);
        }
    };

    // AWS数据获取
    const fetchAwsData = async () => {
        try {
            setLoading(true);
            const promises = instances.aws.map(async (instance) => {
                try {
                    const response = await fetch(`/awsrds/${instance}`);
                    const data = await response.json();
                    
                    if (!data.error) {
                        return {
                            key: instance,
                            env: instance,
                            snapshot_id: data.snapshot_id,
                            snapshot_arn: data.snapshot_arn,
                            snapshot_create_time: data.snapshot_create_time,
                            status: data.status,
                            instance_id: data.instance_id,
                            region: data.region,
                            export_task_id: data.export_task_id
                        };
                    } else {
                        console.error(`获取实例 ${instance} 数据失败:`, data.error);
                        return null;
                    }
                } catch (error) {
                    console.error(`获取实例 ${instance} 数据出错:`, error);
                    return null;
                }
            });

            const results = await Promise.all(promises);
            setAwsData(results.filter(item => item !== null));
        } catch (error) {
            console.error('获取AWS快照列表出错：', error);
            message.error('获取AWS快照列表失败');
        } finally {
            setLoading(false);
        }
    };

    // 导出到S3
    const handleExportToS3 = async (record) => {
        try {
            message.loading('正在启动导出任务...', 2);
            const response = await fetch(`/alirds/export/s3/${record.env}`, {
                method: 'POST'
            });
            const data = await response.json();
            
            if (!data.error) {
                message.success('导出任务已启动');
                message.info(`备份将上传至 ${data.s3_bucket} (${data.region})`);
            } else {
                message.error('导出失败：' + data.error);
            }
        } catch (error) {
            console.error('导出到S3出错：', error);
            message.error('导出失败');
        }
    };

    // 初始化加载
    React.useEffect(() => {
        fetchInstances();
        fetchS3Config();
    }, []);

    // 当实例列表或标签页变化时获取数据
    React.useEffect(() => {
        if (instances.aliyun.length > 0 || instances.aws.length > 0) {
            if (activeTab === '1') {
                fetchAliData();
            } else {
                fetchAwsData();
            }
        }
    }, [activeTab, instances]);

    // 表格列定义优化
    const aliColumns = [
        {
            title: '环境',
            dataIndex: 'env',
            key: 'env',
            render: (text) => <Tag color="blue">{text}</Tag>,
            sorter: (a, b) => a.env.localeCompare(b.env),  // 添加排序
            filters: instances.aliyun.map(env => ({  // 添加筛选
                text: env,
                value: env,
            })),
            onFilter: (value, record) => record.env === value,
        },
        {
            title: '备份开始时间',
            dataIndex: 'backup_start_time',
            key: 'backup_start_time',
            render: (text) => text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-',
            sorter: (a, b) => new Date(a.backup_start_time) - new Date(b.backup_start_time),  // 添加时间排序
            defaultSortOrder: 'descend',  // 默认按时间降序
        },
        {
            title: '下载链接',
            dataIndex: 'backup_download_url',
            key: 'backup_download_url',
            render: (text) => text ? (
                <Button type="link" onClick={() => window.open(text)}>
                    下载
                </Button>
            ) : '-'
        },
        {
            title: '内网下载链接',
            dataIndex: 'backup_intranet_download_url',
            key: 'backup_intranet_download_url',
            render: (text) => text ? (
                <Button type="link" onClick={() => window.open(text)}>
                    下载(内网)
                </Button>
            ) : '-'
        },
        {
            title: '操作',
            key: 'action',
            render: (_, record) => (
                <Space>
                    <Tooltip title={!s3Config ? 'S3配置未加载' : '导出到S3'}>
                        <Button 
                            type="primary"
                            icon={<CloudUploadOutlined />}
                            onClick={() => {
                                Modal.confirm({  // 添加确认对话框
                                    title: '确认导出到S3',
                                    content: (
                                        <div>
                                            <p>确定要将以下备份导出到S3吗？</p>
                                            <ul>
                                                <li>环境：{record.env}</li>
                                                <li>备份时间：{dayjs(record.backup_start_time).format('YYYY-MM-DD HH:mm:ss')}</li>
                                                <li>目标存储桶：{s3Config?.bucket_name}</li>
                                                <li>目标区域：{s3Config?.region}</li>
                                            </ul>
                                            <Alert 
                                                message="导出过程可能需要较长时间，请耐心等待" 
                                                type="warning" 
                                                showIcon 
                                            />
                                        </div>
                                    ),
                                    okText: '确认导出',
                                    cancelText: '取消',
                                    onOk: () => handleExportToS3(record)
                                });
                            }}
                            disabled={!s3Config}
                        >
                            导出到S3
                        </Button>
                    </Tooltip>
                </Space>
            )
        }
    ];

    const awsColumns = [
        {
            title: '环境',
            dataIndex: 'env',
            key: 'env',
            render: (text) => <Tag color="blue">{text}</Tag>,
            sorter: (a, b) => a.env.localeCompare(b.env),
            filters: instances.aws.map(env => ({
                text: env,
                value: env,
            })),
            onFilter: (value, record) => record.env === value,
        },
        {
            title: '快照ID',
            dataIndex: 'snapshot_id',
            key: 'snapshot_id',
        },
        {
            title: '创建时间',
            dataIndex: 'snapshot_create_time',
            key: 'snapshot_create_time',
            render: (text) => text ? dayjs(text).format('YYYY-MM-DD HH:mm:ss') : '-'
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (status) => (
                <Tag color={status === 'available' ? 'green' : 'processing'}>
                    {status}
                </Tag>
            )
        },
        {
            title: '区域',
            dataIndex: 'region',
            key: 'region',
        },
        {
            title: '导出状态',
            dataIndex: 'export_status',
            key: 'export_status',
            render: (status) => {
                if (!status) return '-';
                const statusColors = {
                    'STARTING': 'processing',
                    'IN_PROGRESS': 'processing',
                    'COMPLETE': 'success',
                    'FAILED': 'error',
                    'CANCELING': 'warning',
                    'CANCELED': 'default'
                };
                return (
                    <Tag color={statusColors[status] || 'default'}>
                        {status}
                    </Tag>
                );
            }
        },
        {
            title: '操作',
            key: 'action',
            render: (_, record) => (
                <Space>
                    <Tooltip title={record.status !== 'available' ? '快照状态不可用' : '导出快照'}>
                        <Button
                            type="primary"
                            icon={<CloudUploadOutlined />}
                            onClick={() => handleExportAwsSnapshot(record)}
                            disabled={record.status !== 'available'}
                        >
                            导出快照
                        </Button>
                    </Tooltip>
                </Space>
            )
        }
    ];

    // 添加轮询函数
    const startPolling = () => {
        if (!polling) {
            setPolling(true);
            const pollInterval = setInterval(() => {
                if (activeTab === '2') {  // 只在AWS标签页激活时轮询
                    fetchAwsData();
                }
            }, 30000);  // 每30秒轮询一次

            // 清理函数
            return () => {
                clearInterval(pollInterval);
                setPolling(false);
            };
        }
    };

    // 修改 AWS 快照导出函数
    const handleExportAwsSnapshot = async (record) => {
        try {
            // 显示确认对话框
            Modal.confirm({
                title: '确认导出AWS RDS快照',
                content: (
                    <div>
                        <p>请确认以下快照导出信息：</p>
                        <ul style={{ paddingLeft: '20px' }}>
                            <li>环境：{record.env}</li>
                            <li>快照ID：{record.snapshot_id}</li>
                            <li>创建时间：{dayjs(record.snapshot_create_time).format('YYYY-MM-DD HH:mm:ss')}</li>
                            <li>数据库引擎：{record.engine} {record.engine_version}</li>
                            <li>实例ID：{record.instance_id}</li>
                            <li>区域：{record.region}</li>
                        </ul>
                        <Alert
                            message="注意事项"
                            description={
                                <ul style={{ paddingLeft: '20px', marginBottom: 0 }}>
                                    <li>导出过程可能需要较长时间，请耐心等待</li>
                                    <li>导出期间请勿关闭页面</li>
                                    <li>导出完成后会自动刷新数据</li>
                                </ul>
                            }
                            type="warning"
                            showIcon
                            style={{ marginTop: '10px' }}
                        />
                    </div>
                ),
                okText: '确认导出',
                cancelText: '取消',
                width: 550,
                onOk: async () => {
                    message.loading('正在启动快照导出...', 2);
                    const response = await fetch(`/awsrds/export/${record.env}`, {
                        method: 'POST'
                    });
                    const data = await response.json();
                    
                    if (!data.error) {
                        message.success('快照导出任务已启动');
                        Modal.success({
                            title: '导出任务已启动',
                            content: (
                                <div>
                                    <p>导出任务已成功启动，请注意以下信息：</p>
                                    <ul style={{ paddingLeft: '20px' }}>
                                        <li>导出任务ID：{data.export_task_id}</li>
                                        <li>目标区域：{data.target_region || record.region}</li>
                                        <li>预计完成时间：{data.estimated_completion_time || '未知'}</li>
                                    </ul>
                                    <p>系统将自动轮询任务状态，您可以通过刷新按钮查看最新进度。</p>
                                </div>
                            ),
                            width: 500,
                        });
                        startPolling();  // 启动轮询
                    } else {
                        Modal.error({
                            title: '导出失败',
                            content: (
                                <div>
                                    <p>快照导出失败，错误信息：</p>
                                    <p style={{ color: '#ff4d4f' }}>{data.error}</p>
                                    <p>请检查以下内容：</p>
                                    <ul style={{ paddingLeft: '20px' }}>
                                        <li>确保快照状态为 "available"</li>
                                        <li>确保有足够的存储空间</li>
                                        <li>检查导出权限配置</li>
                                    </ul>
                                </div>
                            ),
                            width: 500,
                        });
                    }
                },
                onCancel() {
                    message.info('已取消导出操作');
                },
            });
        } catch (error) {
            console.error('导出快照出错：', error);
            Modal.error({
                title: '导出失败',
                content: '导出过程发生错误，请稍后重试',
            });
        }
    };

    // 在组件卸载时清理轮询
    React.useEffect(() => {
        return () => {
            if (polling) {
                setPolling(false);
            }
        };
    }, []);

    // 添加表格配置
    const tableProps = {
        pagination: {  // 分页配置
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
            defaultPageSize: 10,
            pageSizeOptions: ['10', '20', '50']
        },
        scroll: { x: 'max-content' },  // 表格横向滚动
        size: 'middle',
        bordered: true,
        rowClassName: (record) => {  // 根据状态设置行样式
            if (record.status === 'FAILED') return 'table-row-failed';
            if (record.status === 'IN_PROGRESS') return 'table-row-processing';
            return '';
        }
    };

    // 添加样式
    const styles = `
        .table-row-failed {
            background-color: #fff1f0;
        }
        .table-row-processing {
            background-color: #e6f7ff;
        }
        .ant-table-row:hover {
            cursor: pointer;
        }
    `;

    // 在组件中添加样式
    React.useEffect(() => {
        const styleSheet = document.createElement('style');
        styleSheet.innerText = styles;
        document.head.appendChild(styleSheet);
        return () => styleSheet.remove();
    }, []);

    return (
        <div>
            <Tabs 
                activeKey={activeTab} 
                onChange={key => setActiveTab(key)}
                tabBarExtraContent={  // 添加刷新按钮
                    <Button 
                        onClick={() => activeTab === '1' ? fetchAliData() : fetchAwsData()}
                        icon={<ReloadOutlined />}
                    >
                        刷新数据
                    </Button>
                }
            >
                <Tabs.TabPane tab="阿里云RDS备份" key="1">
                    <Card>
                        <Table
                            {...tableProps}
                            columns={aliColumns}
                            dataSource={aliData}
                            loading={loading}
                            rowKey="key"
                        />
                    </Card>
                </Tabs.TabPane>
                <Tabs.TabPane tab="AWS RDS快照" key="2">
                    <Card>
                        <Table
                            {...tableProps}
                            columns={awsColumns}
                            dataSource={awsData}
                            loading={loading}
                            rowKey="key"
                        />
                    </Card>
                </Tabs.TabPane>
            </Tabs>
        </div>
    );
}

// 使用立即执行函数来确保只导出一次
(function() {
    if (window.BackupList) {
        console.log('BackupList already exported, skipping...');
        return;
    }

    window.BackupList = function WrappedBackupList() {
        console.log('Rendering wrapped BackupList');
        try {
            return (
                <ErrorBoundary>
                    <BackupList />
                </ErrorBoundary>
            );
        } catch (error) {
            console.error('Error in WrappedBackupList:', error);
            return <div>Error loading BackupList</div>;
        }
    };
    
    console.log('BackupList exported successfully');
})();

console.log('BackupList component loaded');
