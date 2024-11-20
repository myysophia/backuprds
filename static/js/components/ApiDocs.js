function ApiDocs() {
    return (
        <Layout>
            <Header className="site-header">
                <Title level={3} className="site-title">API 文档</Title>
                <div className="header-actions">
                    <Button 
                        type="ghost" 
                        className="api-doc-button"
                        onClick={() => window.close()}
                    >
                        关闭
                    </Button>
                </div>
            </Header>
            <Content>
                <iframe 
                    src="/doc/index.html" 
                    style={{
                        width: '100%',
                        height: 'calc(100vh - 64px)',
                        border: 'none'
                    }}
                />
            </Content>
        </Layout>
    );
} 