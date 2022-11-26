use std::process::Command;
use clap::Parser;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Argo CD Hostname
    #[arg(short = 'a', long)]
    argocd_host: String,

    /// Argo CD Token
    #[arg(long, env = "ARGONAUT_ARGOCD_TOKEN")]
    argocd_token: String,
    
    /// Argo CD Application to diff
    #[arg(short = 'n', long)]
    app_name: String,

    /// Commit hash to diff against live
    #[arg(short, long, env = "CI_COMMIT_SHA")]
    revision: String,

    /// Gitlab Hostname
    #[arg(short = 'g', long, env = "CI_SERVER_HOST")]
    gitlab_host: String,

    /// Gitlab Project ID
    #[arg(short = 'p', long, env = "CI_PROJECT_ID")]
    gitlab_project_id: String,

    /// Gitlab Merge Request ID
    #[arg(long, short = 'm', long, env = "CI_MERGE_REQUEST_IID")]
    gitlab_mr_id: String,

    /// Gitlab Token
    #[arg(long, env = "ARGONAUT_GITLAB_TOKEN")]
    gitlab_token: String,
}

fn main() {
    println!("Starting Argonaut");

    let args = Args::parse();

    let argocd_hostname = &args.argocd_host;
    let argocd_token = &args.argocd_token;
    let app_name = &args.app_name;
    let revision = &args.revision;

    let diff = get_argocd_diff(argocd_hostname, argocd_token, app_name, revision);
    println!("{diff}");

    let gitlab_hostname = &args.gitlab_host;
    let gitlab_token = &args.gitlab_token;
    let project_id = &args.gitlab_project_id;
    let mr_id = &args.gitlab_mr_id;
    post_gitlab_comment(
        gitlab_token,
        gitlab_hostname,
        project_id,
        mr_id,
        &diff,
    );
}

fn get_argocd_diff(hostname: &str, token: &str, app_name: &str, revision: &str) -> String {
    let cmd_stdout = Command::new("argocd")
        .arg("app")
        .arg("diff")
        .arg("--server")
        .arg(hostname)
        .arg("--auth-token")
        .arg(token)
        .arg(app_name)
        .arg("--revision")
        .arg(revision)
        .output()
        .expect("Failed")
        .stdout;
    let raw_diff = String::from_utf8_lossy(&cmd_stdout).to_string();
    return raw_diff.trim().to_string();
}

fn post_gitlab_comment(
    token: &str,
    gitlab_server: &str,
    project_id: &str,
    mr_id: &str,
    text: &String,
) {
    let endpoint = format!(
        "https://{gitlab_server}/api/v4/projects/{project_id}/merge_requests/{mr_id}/discussions"
    );
    let body = format!(
        r#"
```diff
{text}
```"#
    );
    let client = reqwest::blocking::Client::new();
    let res = client
        .post(&endpoint)
        .header("PRIVATE-TOKEN", token)
        .query(&[("body", body)])
        .send()
        .expect("Failed during request");
    let res_text = res.text().expect("Failed to get text");
    println!("{res_text}")
}
