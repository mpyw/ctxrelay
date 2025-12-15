#!/usr/bin/env perl
#
# verify-test-patterns.pl
#
# Verifies that test pattern numbering is consistent across checkers.
# Patterns with the same number within a group should test the same concept.
#
# Naming convention: XXnn where XX is 2-letter prefix, nn is number
#   Goroutine group (context usage):
#     GO01, GO02... - goroutine checker
#     GE01, GE02... - errgroup checker
#     GW01, GW02... - waitgroup checker
#   Derive group (deriver function calls):
#     DD01, DD02... - goroutinederive (single)
#     DA01, DA02... - goroutinederiveand (AND)
#     DM01, DM02... - goroutinederivemixed (Mixed)
#
# Usage:
#   ./scripts/verify-test-patterns.pl           # Check all patterns
#   ./scripts/verify-test-patterns.pl -v        # Verbose output
#   ./scripts/verify-test-patterns.pl -q        # Quiet mode (exit code only)
#

use strict;
use warnings;
use File::Find;
use Getopt::Std;

my %opts;
getopts('vq', \%opts);
my $verbose = $opts{v} // 0;
my $quiet = $opts{q} // 0;

my $script_dir = $0 =~ s|[^/]+$||r;
my $testdata_dir = "${script_dir}../pkg/analyzer/testdata/src";

# Checker configurations: directory => prefix
# Goroutine group
my %goroutine_checkers = (
    'goroutine' => 'GO',
    'errgroup'  => 'GE',
    'waitgroup' => 'GW',
);

# Derive group
my %derive_checkers = (
    'goroutinederive'      => 'DD',
    'goroutinederiveand'   => 'DA',
    'goroutinederivemixed' => 'DM',
);

# patterns_by_group_num: { "G:01" => [ { prefix => 'GO', desc => '...' }, ... ] }
my %patterns_by_group_num;

# Extract patterns from test files
sub extract_patterns {
    my ($dir, $prefix) = @_;
    return unless -d $dir;

    find(sub {
        return unless /\.go$/;
        open my $fh, '<', $_ or return;
        while (<$fh>) {
            # Match: // GO01: description
            if (m{//\s*($prefix)(\d+):\s*(.+)}) {
                my ($pfx, $num, $desc) = ($1, $2, $3);
                $desc =~ s/\s+$//;  # trim trailing whitespace

                # Determine group (G for Goroutine, D for Derive)
                my $group = substr($pfx, 0, 1);
                my $key = "$group:$num";

                push @{$patterns_by_group_num{$key}}, { prefix => $pfx, desc => $desc };
            }
        }
        close $fh;
    }, $dir);
}

# Collect patterns from all checkers
for my $checker (keys %goroutine_checkers) {
    my $dir = "$testdata_dir/$checker";
    my $prefix = $goroutine_checkers{$checker};
    extract_patterns($dir, $prefix);
}

for my $checker (keys %derive_checkers) {
    my $dir = "$testdata_dir/$checker";
    my $prefix = $derive_checkers{$checker};
    extract_patterns($dir, $prefix);
}

# Normalize description for comparison
sub normalize {
    my $desc = shift;
    $desc = lc($desc);
    $desc =~ s/^\s+|\s+$//g;
    $desc =~ s/\s+/ /g;
    return $desc;
}

# Check for inconsistencies
my @errors;

for my $key (sort keys %patterns_by_group_num) {
    my @entries = @{$patterns_by_group_num{$key}};
    next if @entries < 2;  # Only check patterns in multiple checkers

    # Group by normalized description
    my %by_desc;
    for my $e (@entries) {
        my $norm = normalize($e->{desc});
        push @{$by_desc{$norm}}, $e;
    }

    if (keys %by_desc > 1) {
        my ($group, $num) = split /:/, $key;
        my $group_name = $group eq 'G' ? 'Goroutine' : 'Derive';
        my $msg = "Pattern #$num ($group_name group) has inconsistent descriptions:\n";
        for my $e (@entries) {
            $msg .= "  $e->{prefix}$num: $e->{desc}\n";
        }
        push @errors, $msg;
    }
}

# Output results
unless ($quiet) {
    if ($verbose) {
        print "=== Goroutine group patterns ===\n";
        for my $key (sort keys %patterns_by_group_num) {
            next unless $key =~ /^G:/;
            my ($group, $num) = split /:/, $key;
            print "Pattern #$num:\n";
            for my $e (@{$patterns_by_group_num{$key}}) {
                print "  $e->{prefix}$num: $e->{desc}\n";
            }
            print "\n";
        }

        print "=== Derive group patterns ===\n";
        for my $key (sort keys %patterns_by_group_num) {
            next unless $key =~ /^D:/;
            my ($group, $num) = split /:/, $key;
            print "Pattern #$num:\n";
            for my $e (@{$patterns_by_group_num{$key}}) {
                print "  $e->{prefix}$num: $e->{desc}\n";
            }
            print "\n";
        }
    }

    if (@errors) {
        print "=== Inconsistencies found ===\n";
        print $_ for @errors;
        print "Found " . scalar(@errors) . " inconsistent pattern(s).\n";
    } else {
        print "All patterns are consistent.\n";
    }
}

exit scalar(@errors);
