const { Header, Content } = Layout;
const { Title } = Typography;

function App() {
    return (
        <Layout>
            <Header style={{ 
                background: '#fff', 
                padding: '0 20px',
                textAlign: 'center'
            }}>
                <Title 
                    level={3} 
                    style={{ 
                        margin: '16px 0',
                        fontWeight: 'bold'
                    }}
                >
                    Nova RDS 备份下载
                </Title>
            </Header>
            <Content style={{ padding: '20px' }}>
                <BackupList />
            </Content>
        </Layout>
    );
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<App />);
