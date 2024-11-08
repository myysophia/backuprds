const { Search } = Input;

function BackupList() {
    const [backups, setBackups] = React.useState([]);
    const [loading, setLoading] = React.useState(false);
    const [searchText, setSearchText] = React.useState('');

    const fetchBackup = async (env) => {
        try {
            setLoading(true);
            const response = await fetch(`/alirds/${env}`);
            const data = await response.json();
            
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
            message.error(`请求失败: ${error.message}`);
        } finally {
            setLoading(false);
        }
    };

    React.useEffect(() => {
        const environments = [
            'vnnox-uat', 'vnnox-cn-db', 'vnnox-sg-db', 'care-eu-db', 'vnnox-eu-db', 'care-us-db',
            'vnnox-us-db'
        ];
        setBackups([]); // 清空现有数据
        environments.forEach(env => fetchBackup(env));
    }, []);

    const columns = [
        {
            title: '环境名称',
            dataIndex: 'env',
            key: 'env',
            sorter: (a, b) => a.env.localeCompare(b.env),
        },
        {
            title: '最后一次备份时间',
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

    const filteredBackups = backups.filter(backup => 
        backup.env.toLowerCase().includes(searchText.toLowerCase())
    );

    return (
        <Card>
            <Search
                placeholder="搜索环境名称"
                allowClear
                enterButton
                style={{ marginBottom: 20 }}
                onChange={(e) => setSearchText(e.target.value)}
            />
            <Spin spinning={loading}>
                <Table
                    columns={columns}
                    dataSource={filteredBackups}
                    pagination={false}
                    rowKey="env"
                />
            </Spin>
        </Card>
    );
}
